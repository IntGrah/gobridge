package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const token = "MTI3NTk1MjU5NzgxMTA2ODkyOQ.G4mqqe.gSJZ14QgmkcanRC7UZTipo_EPyo75AaPridHjU"

func main() {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.Contains(strings.ToLower(m.Content), "john") {
		return
	}

	var responses = []string{
		"Better be seen at oxford than caught at john's",
		"rather go to oxford than st johns tbh",
		"watch your mouth before you get sent to St Johns",
		// "Twice doth Trinity's clock chime, and many a few at St John's rue the hour, sleepless by the May Ball's delight.",
		"johns kinda like a prison fr",
		"when do we demolish st johns again?",
		// "No Cantabrigian wind howleth fiercer than mine antipathy for the folly that is St John's.",
	}

	var response = responses[rand.Intn(len(responses))]

	s.ChannelMessageSendReply(m.ChannelID, response, m.Reference())
}
