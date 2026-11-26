package handlers

import (
	"testing"

	"github.com/hoshinonyaruko/gensokyo/callapi"
)

func TestBuildCard(t *testing.T) {
	msg := callapi.ActionMessage{Params: callapi.ParamsContent{
		Title:       "维护公告",
		Description: "今晚维护",
		URL:         "https://example.com/notice",
		PicURL:      "https://example.com/banner.png",
	}}
	card, err := buildCard(msg)
	if err != nil {
		t.Fatalf("valid card rejected: %v", err)
	}
	if card.Type != "tuwen" || card.Content.URL != "https://example.com/notice" {
		t.Fatalf("unexpected card: %+v", card)
	}

	msg.Params.URL = "http://example.com/notice"
	if _, err := buildCard(msg); err == nil {
		t.Fatal("http url should be rejected")
	}
}
