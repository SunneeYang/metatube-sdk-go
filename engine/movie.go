package engine

import (
	"fmt"
	"sync"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/priority"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
	"gorm.io/gorm/clause"
)

func (e *Engine) searchMovie(keyword string, provider javtube.MovieProvider, lazy bool) ([]*model.MovieSearchResult, error) {
	// Query DB first (by number).
	if info := new(model.MovieInfo); lazy {
		if result := e.db.Where("number = ? AND provider = ?", keyword, provider.Name()).
			First(info); result.Error == nil && info.Valid() /* must be valid */ {
			return []*model.MovieSearchResult{info.ToSearchResult()}, nil
		} // ignore DB query error.
	}
	// Regular keyword searching.
	if searcher, ok := provider.(javtube.MovieSearcher); ok {
		// auto save all search result's metadata
		return searcher.SearchMovie(keyword)
	}
	// Fallback to movie info querying.
	info, err := e.getMovieInfoByID(keyword, provider, true)
	if err != nil {
		return nil, err
	}
	return []*model.MovieSearchResult{info.ToSearchResult()}, nil
}

func (e *Engine) SearchMovie(keyword, name string, lazy bool) ([]*model.MovieSearchResult, error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}
	provider, ok := e.movieProviders[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return e.searchMovie(keyword, provider, lazy)
}

func (e *Engine) searchMovieAll(keyword string) (results []*model.MovieSearchResult, err error) {
	type response struct {
		Results []*model.MovieSearchResult
		Error   error
	}
	respCh := make(chan response)

	var wg sync.WaitGroup
	for _, provider := range e.movieProviders {
		wg.Add(1)
		// Async searching.
		go func(provider javtube.MovieProvider) {
			defer wg.Done()
			results, err := e.searchMovie(keyword, provider, false)
			respCh <- response{
				Results: results,
				Error:   err,
			}
		}(provider)
	}
	go func() {
		wg.Wait()
		// notify when all searching tasks done.
		close(respCh)
	}()

	// response channel.
	for resp := range respCh {
		if resp.Error != nil {
			continue
		}
		results = append(results, resp.Results...)
	}
	if len(results) == 0 {
		err = javtube.ErrNotFound
	}
	return
}

// SearchMovieAll searches the keyword from all providers.
func (e *Engine) SearchMovieAll(keyword string, lazy bool) (results []*model.MovieSearchResult, err error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}

	results, err = e.searchMovieAll(keyword)
	if err != nil {
		return nil, err
	}

	var ps = new(priority.Slice[float64, *model.MovieSearchResult])
	for _, result := range results {
		if !result.Valid() {
			continue
		}
		ps.Append(number.Similarity(keyword, result.Number)*
			float64(e.movieProviders[result.Provider].Priority()), result)
	}
	// sort according to priority.
	return ps.Sort().Underlying(), nil
}

func (e *Engine) getMovieInfoByID(id string, provider javtube.MovieProvider, lazy bool) (info *model.MovieInfo, err error) {
	// Query DB first (by id).
	if info = new(model.MovieInfo); lazy {
		if result := e.db.Where("id = ? AND provider = ?", id, provider.Name()).
			First(info); result.Error == nil && info.Valid() {
			return
		} // ignore DB query error.
	}
	// delayed info auto-save.
	defer func() {
		if err == nil && info.Valid() {
			e.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(info) // ignore error
		}
	}()
	return provider.GetMovieInfoByID(id)
}

func (e *Engine) GetMovieInfoByID(id, name string, lazy bool) (info *model.MovieInfo, err error) {
	provider, ok := e.movieProviders[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return e.getMovieInfoByID(id, provider, lazy)
}
