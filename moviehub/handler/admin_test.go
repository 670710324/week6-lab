package handler

import (
	"strings"
	"testing"

	"moviehub/store"
)

func TestValidateMovieNormalizesOptionalFields(t *testing.T) {
	input := store.MovieInput{
		Title:      "  Example Movie  ",
		PosterPath: stringPointer(" /poster.jpg "),
	}
	if code, message := validateMovie(&input); code != "" {
		t.Fatalf("validateMovie() = %q, %q", code, message)
	}
	if input.Title != "Example Movie" || input.OriginalTitle != "Example Movie" || input.OriginalLanguage != "en" {
		t.Fatalf("movie defaults were not normalized: %+v", input)
	}
	if input.PosterPath == nil || *input.PosterPath != "/poster.jpg" {
		t.Fatalf("poster path = %v, want trimmed path", input.PosterPath)
	}
}

func TestValidateMovieRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name string
		in   store.MovieInput
		code string
	}{
		{
			name: "empty title",
			in:   store.MovieInput{Title: "  "},
			code: "BAD_TITLE",
		},
		{
			name: "long title",
			in:   store.MovieInput{Title: strings.Repeat("x", 256)},
			code: "BAD_TITLE",
		},
		{
			name: "malformed poster URL",
			in:   store.MovieInput{Title: "Movie", PosterPath: stringPointer("httpnot-a-url")},
			code: "BAD_POSTER",
		},
		{
			name: "invalid date",
			in:   store.MovieInput{Title: "Movie", ReleaseDate: stringPointer("2026-02-30")},
			code: "BAD_DATE",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if code, _ := validateMovie(&test.in); code != test.code {
				t.Fatalf("validateMovie() code = %q, want %q", code, test.code)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
