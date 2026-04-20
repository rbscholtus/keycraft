package api

import (
	"errors"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	return NewService(NewSessionManager(testConfig(t)))
}

func TestServiceOpenCloseSession(t *testing.T) {
	svc := newTestService(t)
	sess, err := svc.OpenSession()
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	if sess.ID == "" {
		t.Fatal("empty session ID")
	}
	if err := svc.CloseSession(sess.ID); err != nil {
		t.Fatalf("CloseSession: %v", err)
	}
}

func TestServiceListCorpora(t *testing.T) {
	svc := newTestService(t)
	got, err := svc.ListCorpora()
	if err != nil {
		// data/corpus has only .json caches — ListCorpora filters to .txt,
		// which may be empty. We still want the call to succeed.
		t.Fatalf("ListCorpora: %v", err)
	}
	_ = got // count depends on checkout; we only assert no error
}

func TestServiceListLayouts(t *testing.T) {
	svc := newTestService(t)
	got, err := svc.ListLayouts()
	if err != nil {
		t.Fatalf("ListLayouts: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one .klf layout in data/layouts/")
	}
}

func TestServiceGetLayout(t *testing.T) {
	svc := newTestService(t)
	dto, err := svc.GetLayout("qwerty")
	if err != nil {
		t.Fatalf("GetLayout: %v", err)
	}
	if dto.Name != "qwerty" {
		t.Errorf("name: got %q, want %q", dto.Name, "qwerty")
	}
	if len(dto.Rows) != 3 {
		t.Fatalf("rows: got %d, want 3", len(dto.Rows))
	}
	for i, r := range dto.Rows {
		// Each row is 12 UTF-8 chars; rune-count is 12.
		runes := []rune(r)
		if len(runes) != 12 {
			t.Errorf("row %d: got %d runes, want 12", i, len(runes))
		}
	}
}

func TestServiceGetLayoutMissing(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.GetLayout("definitely-not-a-layout")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetLayout missing: got %v, want ErrNotFound", err)
	}
}

func TestServiceFlipLayout(t *testing.T) {
	svc := newTestService(t)
	orig, err := svc.GetLayout("qwerty")
	if err != nil {
		t.Fatalf("GetLayout: %v", err)
	}
	flipped, err := svc.FlipLayout("qwerty")
	if err != nil {
		t.Fatalf("FlipLayout: %v", err)
	}
	if flipped.Layout.Name != "qwerty-flipped" {
		t.Errorf("name: got %q, want %q", flipped.Layout.Name, "qwerty-flipped")
	}
	// Row 0 should be reversed.
	origRow0 := []rune(orig.Rows[0])
	flipRow0 := []rune(flipped.Layout.Rows[0])
	for i := range 12 {
		if origRow0[i] != flipRow0[11-i] {
			t.Errorf("row 0 mirror broken at %d: orig=%q flipped=%q", i, origRow0[i], flipRow0[11-i])
			break
		}
	}
}

func TestServiceDescribeCorpus(t *testing.T) {
	svc := newTestService(t)
	sess, err := svc.OpenSession()
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	ref := CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0}
	summary, err := svc.DescribeCorpus(sess.ID, ref)
	if err != nil {
		t.Fatalf("DescribeCorpus: %v", err)
	}
	if summary.Name == "" {
		t.Error("corpus name empty")
	}
	if summary.TotalUnigramsCount == 0 {
		t.Error("expected non-zero total unigrams")
	}
}

func TestServiceDescribeCorpusBadSession(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.DescribeCorpus("nope", CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0})
	if err != ErrSessionNotFound {
		t.Fatalf("got %v, want ErrSessionNotFound", err)
	}
}

func TestServiceAnalyse(t *testing.T) {
	svc := newTestService(t)
	sess, err := svc.OpenSession()
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	resp, err := svc.Analyse(sess.ID, AnalyseRequest{
		Layouts: []string{"qwerty", "colemak-dh"},
		Corpus:  CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("got %d results, want 2", len(resp.Results))
	}
	if resp.Results[0].Name != "qwerty" {
		t.Errorf("result[0].Name: got %q, want qwerty", resp.Results[0].Name)
	}
	if _, ok := resp.Results[0].Metrics["SFB"]; !ok {
		t.Error("expected SFB metric present")
	}
}

func TestServiceAnalyseEmptyLayouts(t *testing.T) {
	svc := newTestService(t)
	sess, _ := svc.OpenSession()
	_, err := svc.Analyse(sess.ID, AnalyseRequest{
		Corpus: CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}

func TestServiceRank(t *testing.T) {
	svc := newTestService(t)
	sess, _ := svc.OpenSession()
	resp, err := svc.Rank(sess.ID, RankRequest{
		Layouts: []string{"qwerty", "colemak-dh"},
		Corpus:  CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if err != nil {
		t.Fatalf("Rank: %v", err)
	}
	if len(resp.Scores) != 2 {
		t.Fatalf("scores: got %d, want 2", len(resp.Scores))
	}
	if len(resp.Medians) == 0 {
		t.Error("expected medians populated")
	}
}
