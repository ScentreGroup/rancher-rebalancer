package evencattle

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/urfave/cli"
	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/slack-incoming-webhook-tools/util"
)

func NotifySlack(c *cli.Context, msg string) {
	if len(c.String("webhook-url")) < 1 {
		log.Debugf("No webhook url provided so no sending notification to slack")
		return
	}

	url := c.String("webhook-url")
	var payload []byte

	// use default message
	if len(c.String("payload")) < 1 && len(c.String("template")) < 1 {
		payload = []byte(msg)
	} else if len(c.String("payload")) >= 1{
		payload = []byte(c.String("payload"))
	} else {
		// get environment variables to supply to the template
		env := util.ReadEnv()
		log.Debug("env: ", env)

		// load template
		t, err := util.Parse(c.String("template"))
		if err != nil {
			log.Fatal(err)
		}

		// render the template
		var tpl bytes.Buffer
		if err := t.Execute(&tpl, env); err != nil {
			log.Fatal(err)
		}
		log.Debug("rendered: ", tpl.String())

		payload = []byte(tpl.String())
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	log.WithFields(log.Fields{
		"payload": string(payload),
		"headers": req.Header,
	}).Debug("request")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"status":  resp.Status,
		"headers": resp.Header,
		"body":    string(body),
	}).Debug("response")

	return
}
