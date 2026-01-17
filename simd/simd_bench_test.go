package main

import (
	"encoding/json"
	"mcompiler/arena"
	"testing"

	"github.com/valyala/fastjson"
)

var jsonBytes = []byte(`{
	"statuses": [
		{
			"metadata": { "result_type": "recent", "iso_language_code": "ja" },
			"created_at": "Sun Aug 31 00:29:15 +0000 2014",
			"id": 505874924095815681,
			"id_str": "505874924095815681",
			"text": "Slack is awesome. Using arena allocator makes it faster.",
			"source": "<a href=\"http://twitter.com/download/iphone\" rel=\"nofollow\">Twitter for iPhone</a>",
			"truncated": false,
			"user": {
				"id": 2244994945,
				"name": "Mobi Dick",
				"screen_name": "mobidick",
				"location": "Seoul",
				"description": "Just a whale.",
				"url": "https://example.com",
				"followers_count": 142,
				"friends_count": 1833,
				"favourites_count": 10245,
				"is_translator": false
			},
			"retweet_count": 0,
			"favorite_count": 0,
			"favorited": false,
			"retweeted": false,
			"lang": "en",
			"coordinates": null,
			"bg_color": null,
			"use_bg": true
		}
	]
}`)

var largeJsonBytes []byte
var deepNestJSON []byte
var floatArrayJSON []byte

func init() {
	// Construct a large JSON (~1MB) by repeating the tweet object
	tweetStr := `{
			"metadata": { "result_type": "recent", "iso_language_code": "ja" },
			"created_at": "Sun Aug 31 00:29:15 +0000 2014",
			"id": 505874924095815681,
			"id_str": "505874924095815681",
			"text": "Slack is awesome. Using arena allocator makes it faster.",
			"source": "<a href=\"http://twitter.com/download/iphone\" rel=\"nofollow\">Twitter for iPhone</a>",
			"truncated": false,
			"user": {
				"id": 2244994945,
				"name": "Gemini User",
				"screen_name": "gemini_dev",
				"followers_count": 142,
				"friends_count": 1833,
				"favourites_count": 10245,
				"is_translator": false
			},
			"retweet_count": 0,
			"favorite_count": 0,
			"favorited": false,
			"retweeted": false,
			"lang": "en",
			"coordinates": null
		}`

	// Create ~1000 items -> ~500KB - 1MB range
	largeJsonBytes = append(largeJsonBytes, []byte(`{"statuses": [`)...)
	for i := 0; i < 3000; i++ {
		if i > 0 {
			largeJsonBytes = append(largeJsonBytes, ',')
		}
		largeJsonBytes = append(largeJsonBytes, []byte(tweetStr)...)
	}
	largeJsonBytes = append(largeJsonBytes, []byte("]}")...) // Close array and object

	// Generate deep nest (~1000 levels)
	deepNestJSON = make([]byte, 0, 4000)
	for i := 0; i < 1000; i++ {
		deepNestJSON = append(deepNestJSON, []byte(`{"a":`)...)
	}
	deepNestJSON = append(deepNestJSON, '1')
	for i := 0; i < 1000; i++ {
		deepNestJSON = append(deepNestJSON, '}')
	}

	// Generate float array (~1000 floats)
	floatArrayJSON = make([]byte, 0, 10000)
	floatArrayJSON = append(floatArrayJSON, '[')
	for i := 0; i < 1000; i++ {
		if i > 0 {
			floatArrayJSON = append(floatArrayJSON, ',')
		}
		floatArrayJSON = append(floatArrayJSON, []byte("123.456789")...)
	}
	floatArrayJSON = append(floatArrayJSON, ']')
}

func BenchmarkStdJSON_Map(b *testing.B) {
	for b.Loop() {
		var m map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &m); err != nil {
			b.Fatal(err)
		}
	}
}

type Tweet struct {
	Statuses []struct {
		Text string `json:"text"`
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
	} `json:"statuses"`
}

func BenchmarkStdJSON_Struct(b *testing.B) {
	for b.Loop() {
		var t Tweet
		if err := json.Unmarshal(jsonBytes, &t); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFastParser(b *testing.B) {
	a := arena.NewBestArena()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Reset()

		p := NewParser(jsonBytes, a)
		_ = p.ParseAny()
	}
}

func BenchmarkStdJSON_Map_Large(b *testing.B) {
	for b.Loop() {
		var m map[string]interface{}
		if err := json.Unmarshal(largeJsonBytes, &m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdJSON_Struct_Large(b *testing.B) {
	for b.Loop() {
		var t Tweet
		if err := json.Unmarshal(largeJsonBytes, &t); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFastParser_Large(b *testing.B) {
	a := arena.NewBestArena()
	b.ResetTimer()
	for b.Loop() {
		a.Reset()
		p := NewParser(largeJsonBytes, a)
		_ = p.ParseAny()
	}
}

func BenchmarkValyala_Small(b *testing.B) {
	var p fastjson.Parser
	b.ResetTimer()
	for b.Loop() {
		_, err := p.ParseBytes(jsonBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValyala_Large(b *testing.B) {
	var p fastjson.Parser
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseBytes(largeJsonBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFastParser_DeepNest(b *testing.B) {
	a := arena.NewBestArena()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Reset()
		p := NewParser(deepNestJSON, a)
		_ = p.ParseAny()
	}
}

func BenchmarkFastParser_FloatArray(b *testing.B) {
	a := arena.NewBestArena()
	b.ResetTimer()
	for b.Loop() {
		a.Reset()
		p := NewParser(floatArrayJSON, a)
		_ = p.ParseAny()
	}
}

func BenchmarkStd_DeepNest(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		var res interface{}
		_ = json.Unmarshal(deepNestJSON, &res)
	}
}

func BenchmarkStd_FloatArray(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		var res interface{}
		_ = json.Unmarshal(floatArrayJSON, &res)
	}
}

func BenchmarkValyala_DeepNest(b *testing.B) {
	b.Skip("fastjson is not able to parse deep nested JSON")
	var p fastjson.Parser
	b.ResetTimer()
	for b.Loop() {
		_, err := p.ParseBytes(deepNestJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValyala_FloatArray(b *testing.B) {
	var p fastjson.Parser
	b.ResetTimer()
	for b.Loop() {
		_, err := p.ParseBytes(floatArrayJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --------------------------------------------------------------------------
// String Scanning Micro-Benchmarks
// --------------------------------------------------------------------------

// 1. String with escapes (Worst case for naive scanning, best for SIMD single-pass)
// 100 chars, escape at the end
var stringBenchData = []byte(`"This is a relatively long string that has an escaped quote \" right here to test the scanning logic."`)

func BenchmarkScanStringBoundary(b *testing.B) {
	a := arena.NewBestArena()
	p := NewParser(stringBenchData, a)
	p.cursor = 1

	b.ResetTimer()
	for b.Loop() {
		p.cursor = 1
		_, _ = p.scanStringBoundary()
	}
}
