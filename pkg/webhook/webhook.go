package webhook

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

const defaultWxWorkTpl = `
	{
		"msgtype": "text",
		"text": {
			"content": "{{ .Message }}"
		}
	}
`

type TplWebHook struct {
	webhook string
	tpl     *template.Template
}

type TplData struct {
	Message string
}

func NewTplWebHook(webhook string, webhookTpl string) *TplWebHook {
	if webhookTpl == "" {
		webhookTpl = defaultWxWorkTpl
	}
	tpl, err := template.New("webhook").Parse(webhookTpl)
	if err != nil {
		log.Fatalf("创建 Webhook 渲染模版异常: %s", err.Error())
	}
	return &TplWebHook{
		webhook: webhook,
		tpl:     tpl,
	}
}

func (n *TplWebHook) SendHook(message string) {
	go func() {
		// 发送文本消息
		data := TplData{
			Message: message,
		}

		var buf bytes.Buffer
		err := n.tpl.Execute(&buf, data)
		if err != nil {
			log.Errorf("渲染 Webhook 消息异常: %s", err.Error())
			return
		}

		body := strings.NewReader(buf.String())
		req, _ := http.NewRequest("POST", n.webhook, body)
		req.Header.Add("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf(err.Error())
		}
		if resp.StatusCode != 200 {
			raw, _ := io.ReadAll(resp.Body)
			log.Errorf(string(raw))
		}
	}()
}
