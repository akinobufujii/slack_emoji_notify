package main

import (
	"io/ioutil"
	"strings"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackutilsx"
	yaml "gopkg.in/yaml.v2"
)

// SettingData この通知に必要な設定データ
type SettingData struct {
	AccessToken        string `yaml:"access_token"`
	BotUserAccessToken string `yaml:"bot_user_access_token"`
	TargetChannelID    string `yaml:"target_channel_id"`
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
	accessAPI := slack.New(settingData.AccessToken)
	botUserAccessAPI := slack.New(settingData.BotUserAccessToken)

	// WebSocketでSlack RTM APIに接続する
	rtm := botUserAccessAPI.NewRTM()

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

					if len(textSlice) <= 1 {
						rtm.SendMessage(rtm.NewOutgoingMessage("私は以下の言葉に反応します\n`ping`\n`url`", ev.Channel))
					}

					switch textSlice[1] {
					case "ping":
						rtm.SendMessage(rtm.NewOutgoingMessage("私は生きてます", ev.Channel))
						break

					case "url":
						emojiMap, err := accessAPI.GetEmoji()
						if err != nil {
							panic(err)
						}

						if len(textSlice) <= 2 {
							rtm.SendMessage(rtm.NewOutgoingMessage("引数が足りません\n使い方は `url 「任意の絵文字」` です", ev.Channel))
							continue
						}

						emojiKey := strings.Replace(textSlice[2], ":", "", -1)

						value, ok := emojiMap[emojiKey]
						if ok {
							// ※エイリアス未対応
							messageStr := "絵文字のurlは以下です\n" + value
							rtm.SendMessage(rtm.NewOutgoingMessage(messageStr, ev.Channel))
						} else {
							rtm.SendMessage(rtm.NewOutgoingMessage("絵文字のurlの獲得に失敗しました", ev.Channel))
						}
					}
					break
				}
			}
			break
		}
	}
}
