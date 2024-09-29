package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/go-chi/render"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/anthropic"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const embeddingModelName = "text-embedding-ada-002"

func setupWeave(_ context.Context, cfg config) (*weaviate.Store, error) {
	openaiClient, err := openai.New(
		openai.WithModel("gpt-3.5-turbo-0125"),
		openai.WithEmbeddingModel(embeddingModelName),
	)
	if err != nil {
		return nil, err
	}
	emb, err := embeddings.NewEmbedder(openaiClient)
	if err != nil {
		return nil, err
	}
	wvStore, err := weaviate.New(
		weaviate.WithEmbedder(emb),
		weaviate.WithScheme("http"),
		weaviate.WithHost(cfg.WeaviateHost+":"+cfg.WeaviatePort),
		weaviate.WithIndexName("Document"),
	)
	if err != nil {
		return nil, err
	}

	return &wvStore, nil
}

func main() {
	ctx := context.Background()
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("problem parsing config: %+v", err)
	}
	logger := setupLogger(ctx, filepath.Join(cfg.LogDir, "server.log"))

	weave, err := setupWeave(ctx, cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to setup weave")
	}

	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-5-sonnet-20240620"),
	)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)                 // add an id to context
	r.Use(middleware.RealIP)                    // do the True-Client-IP, X-Real-IP or the X-Forwarded-For dance
	r.Use(middleware.Logger)                    // log requests
	r.Use(middleware.Recoverer)                 // panic recovery with http 500
	r.Use(middleware.Timeout(60 * time.Second)) // request timeout
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := render.Render(w, r, NewMsgResponse("root")); err != nil {
			_ = render.Render(w, r, ErrRender(err))
			return
		}
	})

	ragHandler := RagHandler{
		ctx:    ctx,
		logger: logger,
		weave:  weave,
		claude: llm,
	}

	r.Post("/add", ragHandler.addDocsHandler)
	r.Post("/query", ragHandler.queryHandler)

	addrStr := fmt.Sprintf(":%d", cfg.Port)
	logger.Fatal().Err(http.ListenAndServe(addrStr, r))
}

type MsgResponse struct {
	Message string `json:"message"`
}

func NewMsgResponse(msg string) *MsgResponse {
	return &MsgResponse{Message: msg}
}

func (m MsgResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

type ErrResponse struct {
	Err            error  `json:"-"`               // low-level runtime error
	HTTPStatusCode int    `json:"-"`               // http response status code
	StatusText     string `json:"status"`          // user-level status message
	ErrorText      string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}
