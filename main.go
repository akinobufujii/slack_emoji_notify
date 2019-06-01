package main

import (
	"io/ioutil"
	"strings"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackutilsx"
	yaml "gopkg.in/yaml.v2"
)

type SettingData struct {
	APIToken        string `yaml:"api_token"`
	TargetChannelID string `yaml:"target_channel_id"`
}

func main() {

	// ymlファイルから設定読み込み
	buf, err := ioutil.ReadFile("setting.yml")
	if err != nil {
		panic(err)
	}

	var settingData SettingData
	err = yaml.Unmarshal(buf, &settingData)
	if err != nil {
		panic(err)
	}

	// APIトークンでbot機能を有効化
	var api = slack.New(settingData.APIToken)

	// WebSocketでSlack RTM APIに接続する
	rtm := api.NewRTM()

	// 自身の情報を手に入れる
	auth, err := rtm.AuthTest()
	if err != nil {
		panic(err)
	}

	// goroutineで並列化する
	go rtm.ManageConnection()

	// イベントを取得する
	for msg := range rtm.IncomingEvents {
		// 型swtichで型を比較する
		switch ev := msg.Data.(type) {
		case *slack.EmojiChangedEvent: // 絵文字変更
			{
				messageString := "絵文字に変化があったよ\n"

				switch ev.SubType {
				case "add": // 追加
					messageString += " *" + ev.Name + "* が追加されたよ\n"
					messageString += "アイコン画像はこれだよ" + ev.Value
					break
				case "remove": // 削除

					for _, iconName := range ev.Names {
						messageString += "," + " *" + iconName + "* "
					}
					messageString += "が削除されたよ"
					break

				}

				// エスケープシーケンスを有効化させる
				messageString = slackutilsx.EscapeMessage(messageString)
				rtm.SendMessage(rtm.NewOutgoingMessage(messageString, settingData.TargetChannelID))
			}
			break

		case *slack.MessageEvent:
			{
				if strings.Contains(ev.Text, "<@"+auth.UserID+">") {
					// 自分宛てのメッセージがある
					textSlice := strings.Split(ev.Text, " ")
					for _, text := range textSlice {
						switch text {
						case "ping":
							rtm.SendMessage(rtm.NewOutgoingMessage("私は生きてます", ev.Channel))
							break
						}
					}
				}
			}
			break
		}
	}
}
