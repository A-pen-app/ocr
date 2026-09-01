package store

import (
	"context"
	"slices"
	"testing"

	"github.com/A-pen-app/ai-client/models"
	mqmodels "github.com/A-pen-app/mq/v2/models"
)

// fakeAIClient returns a canned response and records what it was asked.
type fakeAIClient struct {
	resp string
	err  error
	got  models.AIChatMessage
}

func (f *fakeAIClient) Generate(_ context.Context, m models.AIChatMessage, _ models.AIClientOptions) (string, error) {
	f.got = m
	return f.resp, f.err
}

func TestScanCompany(t *testing.T) {
	file := models.InputFile{URL: "https://example.com/a.heic", MimeType: "image/heic"}

	t.Run("兩個欄位都讀到", func(t *testing.T) {
		c := &fakeAIClient{resp: `{"company_name":"聯合醫療器材行","tax_id":"04595257"}`}
		got, err := NewOcrStore(nil, c, nil).ScanCompany(context.Background(), file)
		if err != nil {
			t.Fatal(err)
		}
		if got.CompanyName == nil || *got.CompanyName != "聯合醫療器材行" {
			t.Errorf("company_name = %v", got.CompanyName)
		}
		if got.TaxID == nil || *got.TaxID != "04595257" {
			t.Errorf("tax_id = %v", got.TaxID)
		}
	})

	t.Run("讀不到的欄位是 nil 而不是空字串", func(t *testing.T) {
		c := &fakeAIClient{resp: `{"company_name":"聯合醫療器材行","tax_id":""}`}
		got, err := NewOcrStore(nil, c, nil).ScanCompany(context.Background(), file)
		if err != nil {
			t.Fatal(err)
		}
		if got.TaxID != nil {
			t.Errorf("tax_id = %v, want nil", *got.TaxID)
		}
	})

	t.Run("附件帶著呼叫端給的 MIME，不靠嗅探", func(t *testing.T) {
		c := &fakeAIClient{resp: `{"company_name":"x","tax_id":""}`}
		if _, err := NewOcrStore(nil, c, nil).ScanCompany(context.Background(), file); err != nil {
			t.Fatal(err)
		}
		if len(c.got.Files) != 1 || c.got.Files[0].MimeType != "image/heic" {
			t.Errorf("files = %+v", c.got.Files)
		}
		if len(c.got.ImageUrls) != 0 {
			t.Errorf("不該走舊的 ImageUrls: %v", c.got.ImageUrls)
		}
	})

	t.Run("模型回非 JSON 就回錯", func(t *testing.T) {
		c := &fakeAIClient{resp: `抱歉，我看不懂這份文件`}
		if _, err := NewOcrStore(nil, c, nil).ScanCompany(context.Background(), file); err == nil {
			t.Error("want error")
		}
	})
}

// fakeMQ records the topics it was asked to publish to.
type fakeMQ struct {
	topics []string
}

func (f *fakeMQ) Send(topic string, _ interface{}, _ ...mqmodels.GetMQOption) error {
	f.topics = append(f.topics, topic)
	return nil
}

func (f *fakeMQ) SendWithContext(_ context.Context, topic string, data interface{}, opts ...mqmodels.GetMQOption) error {
	return f.Send(topic, data, opts...)
}

func (f *fakeMQ) Receive(string) (<-chan []byte, error) { return nil, nil }

func (f *fakeMQ) ReceiveWithContext(context.Context, string) (<-chan []byte, error) {
	return nil, nil
}

func TestScanRawInfoPublishesWhereTheCallerSaid(t *testing.T) {
	cases := []struct {
		name  string
		topic string
		want  []string
	}{
		{"沒設 topic 就不發佈", "", nil},
		{"發到設定的 topic", string(models.OCRTopicDev), []string{"wanderer-dev"}},
		{"跨專案路徑原樣送出", "projects/penpeer-production/topics/wanderer", []string{"projects/penpeer-production/topics/wanderer"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ai := &fakeAIClient{resp: `{"name":"王小明","birthday":"1990-01-01"}`}
			m := &fakeMQ{}

			store := NewOcrStore(m, ai, &Config{MaxToken: 1024, Topic: c.topic})
			if _, err := store.ScanRawInfo(context.Background(), "u1", "https://example.com/licence.jpg", models.PlatformTypeApen); err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(m.topics, c.want) {
				t.Errorf("published to %v, want %v", m.topics, c.want)
			}
		})
	}
}
