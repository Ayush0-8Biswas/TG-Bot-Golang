package Functions

import (
	"Telegram-Bot/Lib/TgTypes"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type SendPhotoResult struct {
	Ok          bool                `json:"ok"`
	Result      TgTypes.MessageType `json:"result"`
	ErrorCode   int                 `json:"error_code"`
	Description string              `json:"description"`
}

type SendPhotoQuery struct {
	ChatId                   int64                       `json:"chat_id"`
	Photo                    string                      `json:"photo"` // Or multipart file
	Caption                  string                      `json:"caption,omitempty"`
	ParseMode                string                      `json:"parse_mode,omitempty"`
	CaptionEntities          []TgTypes.MessageEntityType `json:"caption_entities,omitempty"`
	DisableNotification      bool                        `json:"disable_notification,omitempty"`
	ProtectContent           bool                        `json:"protect_content,omitempty"`
	ReplyToMessageId         int64                       `json:"reply_to_message_id,omitempty"`
	AllowSendingWithoutReply bool                        `json:"allow_sending_without_reply,omitempty"`
	//ReplyMarkup              InlineKeyboardMarkupType `json:"reply_markup,omitempty"`
}

func SendPhotoByReader(baseUrl string, photoPath *bytes.Buffer, message *TgTypes.MessageType, caption string, isProtected bool) (*TgTypes.MessageType, error) {
	client := &http.Client{Timeout: time.Minute * 10}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	sendQuery := make(map[string]interface{})
	sendQuery["chat_id"], sendQuery["reply_to_message_id"], sendQuery["caption"], sendQuery["protect_content"] = message.Chat.Id, message.MessageId, caption, isProtected

	for k, v := range sendQuery {
		fw, err := writer.CreateFormField(k)
		_, err = io.Copy(fw, strings.NewReader(fmt.Sprint(v)))
		if err != nil {
			return nil, err
		}
	}

	fw, err := writer.CreateFormFile("document", "resizing.png")
	_, err = io.Copy(fw, photoPath)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseUrl+"/sendDocument", bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType()) // Very, very important step
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	sendResult, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	returnData := SendPhotoResult{}
	err = json.Unmarshal(sendResult, &returnData)
	if err != nil {
		return nil, err
	}

	if !returnData.Ok {
		return nil, errors.New(returnData.Description)
	}

	return &returnData.Result, nil
}
