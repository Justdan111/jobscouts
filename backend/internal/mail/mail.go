package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"sort"
	"time"

	"jobscout/internal/models"
)

// SendDigest optionally emails the strongest new jobs via Resend IF a key is set
// in the environment. With no key it no-ops — keeping the app free of any
// required third-party service.
func SendDigest(ctx context.Context, jobs []models.Job) (sent int, err error) {
	key := os.Getenv("RESEND_API_KEY")
	to := os.Getenv("DIGEST_TO")
	from := os.Getenv("DIGEST_FROM")
	if from == "" {
		from = "JobScout <onboarding@resend.dev>"
	}

	top := make([]models.Job, 0)
	for _, j := range jobs {
		if j.Status == models.StatusNew && j.Eligibility != models.EligUSOnly {
			top = append(top, j)
		}
	}
	sort.Slice(top, func(i, k int) bool { return top[i].Score > top[k].Score })
	if len(top) > 12 {
		top = top[:12]
	}
	if key == "" || to == "" || len(top) == 0 {
		return 0, nil
	}

	body, _ := json.Marshal(map[string]any{
		"from": from, "to": to,
		"subject": fmt.Sprintf("JobScout: %d new matches", len(top)),
		"html":    render(top),
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+key)
	res, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return 0, fmt.Errorf("resend status %d", res.StatusCode)
	}
	return len(top), nil
}

func render(jobs []models.Job) string {
	var rows bytes.Buffer
	for _, j := range jobs {
		rows.WriteString(fmt.Sprintf(`<tr><td style="padding:12px 0;border-bottom:1px solid #E6E8EC;">
<a href="%s" style="color:#117C6F;font:600 15px sans-serif;text-decoration:none;">%s</a>
<div style="font:400 13px sans-serif;color:#6B7280;">%s · %s · score %d</div></td></tr>`,
			html.EscapeString(j.URL), html.EscapeString(j.Title),
			html.EscapeString(j.Company), html.EscapeString(j.Location), j.Score))
	}
	return `<div style="max-width:560px;margin:0 auto;font-family:sans-serif;"><h1 style="font:700 18px sans-serif;">Today's matches</h1><table style="width:100%;border-collapse:collapse;">` + rows.String() + `</table></div>`
}
