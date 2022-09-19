package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/viper"
)

var logger *MessageLogger
var api *slack.Client
var botID *string

func init() {
	logger = &MessageLogger{debug: true, isatty: isatty.IsTerminal(os.Stdout.Fd())}

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Log("finding home dir", err.Error())
		os.Exit(1)
	}

	// Search config in home directory with name ".robot" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".robot")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	debugConfig, ok := viper.Get("slack_bot_debug").(bool)
	if ok {
		logger.debug = debugConfig
	}
	logger.debug = true

	token, ok := viper.Get("slack_bot_user_oauth_access_token").(string)
	AssertTrue(ok, "SLACK_BOT_USER_OAUTH_ACCESS_TOKEN is required")

	api = slack.New(token)
	// slack.SetLogger(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags))
	// api.SetDebug(false)
}

// BotCmd bot command
type BotCmd struct {
	Channel string
	UserID  string
	Message string
}

// String bot command string
func (b *BotCmd) String() string {
	return fmt.Sprintf("%s %s %s", b.Channel, b.UserID, b.Message)
}

// NewBotCmd instantiates new BotCmd
func NewBotCmd(channel, userID, msg string) BotCmd {
	return BotCmd{
		Channel: channel,
		UserID:  userID,
		Message: msg,
	}
}

func main() {
	cmdQueue := []*BotCmd{}
	users := map[string]string{}

	signingSecret := viper.Get("slack_signing_secret").(string)

	http.HandleFunc("/events-endpoint", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received a new event callback", r.URL)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading body err:", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			logger.Error("new secret verification err:", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, err := sv.Write(body); err != nil {
			logger.Error("secret verification write err:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := sv.Ensure(); err != nil {
			logger.Error("secret verification ensure err:", err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			logger.Error("parsing event err:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			logger.Log("event type url verification", "")

			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			logger.Log("event type callback", "")

			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				logger.Log("app mention event", "")

				userID := ev.User

				logger.Debug("ev", Jsonize(ev))
				logger.Debug("channel", ev.Channel)

				message := RemoveMention(ev.Text)

				logger.Log("message", message)

				cleansedMsg := Cleanse(message)

				slackUsername := GetSlackUsername(users, userID)

				ConfirmRequest(ev.Channel, cleansedMsg, slackUsername)

				botCmd := NewBotCmd(ev.Channel, ev.User, cleansedMsg)

				cmdQueue = append(cmdQueue, &botCmd)
			}
		}

	})

	go func() {
		for {
			time.Sleep(time.Second * 5)
			nextCmd, remainder := Next(cmdQueue)
			if nextCmd == nil {
				continue
			}
			channel := nextCmd.Channel
			logger.Log("executing", nextCmd.Message)
			cmdQueue = remainder

			response, err := Execute(nextCmd.Message)
			if len(response) == 0 && err != nil {
				response = err.Error()
			}
			if len(response) == 0 {
				response = "empty response"
			}
			response = strings.TrimSpace(response)
			SendReponse(channel, response)
			fmt.Println()
		}
	}()

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}

// SendReponse sends response
func SendReponse(channel, response string) {
	_, _, err := api.PostMessage(channel, slack.MsgOptionText(FormatResponse(response), false))
	if err != nil {
		logger.Error("posting msg to channel", err.Error())
	}
}

// GetSlackUsername get slack username
func GetSlackUsername(users map[string]string, userID string) string {
	slackUsername := ""
	if val, ok := users[userID]; ok {
		slackUsername = val
	} else {
		slackUser, err := api.GetUserInfo(userID)
		if err != nil {
			logger.Error("get user info", err.Error())
		}
		slackUsername = slackUser.Name
		users[userID] = slackUsername
	}
	return slackUsername
}

// Cleanse cleanse
func Cleanse(message string) string {
	terms := strings.Split(message, " ")
	cleansed := []string{}
	for _, term := range terms {
		term = RemoveSpecialFormatting(term)
		cleansed = append(cleansed, term)
	}
	cleansedMsg := strings.Join(cleansed, " ")
	logger.Log("cleansed", cleansedMsg)
	return cleansedMsg
}

// Jsonize jsonize
func Jsonize(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		logger.Error("marshalling ev", err.Error())
		return ""
	}
	return string(b)
}

// ConfirmRequest confirm request
func ConfirmRequest(channel, msg, slackUsername string) {
	_, _, err := api.PostMessage(channel, slack.MsgOptionText("executing `"+msg+"` for "+slackUsername, false))
	if err != nil {
		logger.Error("posting message in channel", msg)
	}
}

// Next next
func Next(xs []*BotCmd) (*BotCmd, []*BotCmd) {
	if len(xs) <= 0 {
		return nil, []*BotCmd{}
	}
	i := 0
	y := xs[i]
	ys := append(xs[:i], xs[i+1:]...)
	return y, ys
}

// AssertTrue if the condition is not met, logs assert error and exits
func AssertTrue(condition bool, errmsg string) {
	if !condition {
		logger.Log("error", errmsg)
		os.Exit(1)
	}
}

// Execute parses out command and executes it
func Execute(message string) (string, error) {
	cmd, args := GetCommandAndArgs(message)

	combined := cmd + " " + strings.Join(args, " ")
	logger.Debug("executing", combined)

	output, err := Exec(cmd, args)
	if err != nil {
		logger.Debug("error", err.Error())
		return string(output), err
	}
	return output, nil
}

// Exec executes commands
func Exec(cmd string, args []string) (string, error) {
	command := exec.Command(cmd, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// RemoveMention remove mention
func RemoveMention(text string) string {
	if botID == nil {
		fields := strings.Fields(text)
		for _, s := range fields {
			// is the token a slack mention
			if strings.HasPrefix(s, "<@") && strings.HasSuffix(s, ">") {
				id := s[strings.Index(s, "<@")+2 : len(s)-1]
				logger.Debug("id", id)
				// check to see if the id is bot it
				user, err := api.GetUserInfo(id)
				if err == nil {
					logger.Debug("bot", Jsonize(user))
					if user.IsBot {
						botID = &id
						break
					}
				} else {
					logger.Error("get bot info", err.Error())
				}
			}
		}
		logger.Log("parsed bot id", *botID)
	}
	textAfterMention := strings.Replace(text, "<@"+*botID+"> ", "", -1)
	return textAfterMention
}

// RemoveSpecialFormatting removes special formatting for phone number, URL, etc.
func RemoveSpecialFormatting(message string) string {
	tokens := strings.Fields(message)
	for i, token := range tokens {
		if strings.HasPrefix(token, "<") && strings.HasSuffix(token, ">") && strings.Contains(token, "|") {
			tokens[i] = strings.Replace(strings.Split(token, "|")[1], ">", "", 1)
		}
	}

	return strings.Join(tokens, " ")
}

// GetCommandAndArgs returns command and its arguments
func GetCommandAndArgs(textAfterMention string) (string, []string) {
	tokens := strings.Fields(textAfterMention)

	if len(tokens) == 0 {
		return "", []string{}
	}

	return tokens[0], tokens[1:]
}

// FormatResponse replies back with triple quote
func FormatResponse(response string) string {
	if strings.Contains(response, "http://") || strings.Contains(response, "https://") {
		return response
	}
	return "```" + response + "```"
}
