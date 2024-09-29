package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
	"mime"
	"net/http"
	"strings"
)

type RagHandler struct {
	ctx    context.Context
	logger *zerolog.Logger
	weave  *weaviate.Store
	claude *anthropic.LLM
}

func (rh *RagHandler) addDocsHandler(w http.ResponseWriter, r *http.Request) {

	addDocRequest := AddDocRequest{}
	err := parseRequestJson(r, &addDocRequest)
	if err != nil {
		rh.logger.Error().Err(err).Msg("failed to parse add docs request")
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	var docs []schema.Document
	for _, doc := range addDocRequest.Documents {
		docs = append(docs, schema.Document{
			PageContent: doc.Text,
			Metadata: map[string]interface{}{
				"source": doc.Name,
			},
		})
	}

	_, err = rh.weave.AddDocuments(rh.ctx, docs)
	if err != nil {
		rh.logger.Error().Err(err).Msg("could not add documents")
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	if err := render.Render(w, r, NewMsgResponse("add doc")); err != nil {
		_ = render.Render(w, r, ErrRender(err))
	}
}

func (rh *RagHandler) queryHandler(w http.ResponseWriter, r *http.Request) {
	queryRequest := DocQueryRequest{}
	err := parseRequestJson(r, &queryRequest)
	if err != nil {
		rh.logger.Error().Err(err).Msg("failed to parse query request")
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	similarDocs, err := rh.weave.SimilaritySearch(rh.ctx, queryRequest.Query, 5)
	if err != nil {
		rh.logger.Error().Err(err).Msg("failed to query documents")
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	var docs []string
	for _, doc := range similarDocs {
		docs = append(docs, doc.PageContent)
	}

	claudeQuery := fmt.Sprintf(queryTemplate, queryRequest.Query, strings.Join(docs, "\n"))
	claudeAnswer, err := llms.GenerateFromSinglePrompt(rh.ctx, rh.claude, claudeQuery)
	if err != nil {
		rh.logger.Error().Err(err).Msg("failed to query claude")
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	if err := render.Render(w, r, DocQueryResponse{Text: claudeAnswer}); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

type Doc struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type AddDocRequest struct {
	Documents []Doc `json:"documents"`
}

type DocQueryRequest struct {
	Query string `json:"query"`
}

type DocQueryResponse struct {
	Text string `json:"text"`
}

func (qr DocQueryResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func parseRequestJson(r *http.Request, reqDataType any) error {
	cType := r.Header.Get("Content-Type")
	mType, _, err := mime.ParseMediaType(cType)
	if err != nil {
		return err
	}
	if mType != "application/json" {
		return fmt.Errorf("unexpected mime type: %s", mType)
	}
	jsonDecoder := json.NewDecoder(r.Body)
	jsonDecoder.DisallowUnknownFields()
	err = jsonDecoder.Decode(reqDataType)
	if err != nil {
		return err
	}
	return nil
}

const queryTemplate string = `
You are given a question and a number of pieces of context. 
You may assume the context information has been verified. If the question is
related to the context, answer using the context info. If it is not, the answer 
as you would normally or you may decline to answer saying 'Not enough Info'.

Question:
%s

Context:
%s
`
