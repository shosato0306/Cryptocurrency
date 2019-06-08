package slack

import (
    "bytes"
    "fmt"
	"net/http"
	"cryptocurrency/config"
)

func Notice(channel, text string){
    jsonStr := `{"channel":"` + channel + `","username":"Crypto-bot","text":"` + text + `","icon_emoji":":ghost:"}`

    req, err := http.NewRequest(
        "POST",
        config.Config.SlackWebhookURL,
        bytes.NewBuffer([]byte(jsonStr)),
    )

    if err != nil {
		fmt.Print("Create new request to slack failed: ", err)
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
		fmt.Print("Notificate to slack failed: ", err)
    }

    fmt.Print(resp)
	defer resp.Body.Close()
}
