package alert

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramConfig struct {
	Token       string
	ApiEndPoint string
}

func NewTelegramConfig(token, apiEndPoint string) *TelegramConfig {
	return &TelegramConfig{
		Token:       token,
		ApiEndPoint: apiEndPoint,
	}
}

func SendMesg(config *TelegramConfig, message string, chatID int64) (int, error) {
	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(config.Token, config.ApiEndPoint)
	if err != nil {
		log.Printf("Error creating bot instance: %v", err)
		return 0, err
	}

	// Create a message configuration
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeHTML

	// Send the message
	sentMessage, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return 0, err
	}

	// Return the message ID
	return sentMessage.MessageID, nil
}

func DeleteMesg(config *TelegramConfig, messageID int, chatID int64) error {
	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(config.Token, config.ApiEndPoint)
	if err != nil {
		log.Printf("Error creating bot instance: %v", err)
		return err
	}

	// Create a delete message configuration
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)

	// Send the delete message request
	_, err = bot.Request(deleteMsg)
	if err != nil {
		log.Printf("Error deleting message: %v", err)
		return err
	}

	return nil
}

//func sendMesg
//func deletMesg
