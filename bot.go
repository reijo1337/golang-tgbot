package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	// @BotFather gives you this
	BotToken   = "XXX"
	WebhookURL = "https://525f2cb5.ngrok.io"
	MaxID      = 0
)

type Task struct {
	ID       int
	owner    *tgbotapi.User
	executer *tgbotapi.User
	text     string
}

type TaskManager struct {
	bot   *tgbotapi.BotAPI
	tasks []Task
}

func newBot() (*TaskManager, error) {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return nil, err
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0)

	return &TaskManager{bot, tasks}, nil
}

func (t *TaskManager) execUpdate(update tgbotapi.Update) error {
	user := update.Message.From
	command := update.Message.Text

	if strings.Contains(command, "/new") {
		taskName := command[5:]
		MaxID++
		task := Task{MaxID, user, nil, taskName}
		t.tasks = append(t.tasks, task)
		t.bot.Send(tgbotapi.NewMessage(
			int64(user.ID),
			`Задача "`+taskName+`" создана, id=`+strconv.Itoa(task.ID),
		))
	}
	if strings.Contains(command, "/tasks") {
		if len(t.tasks) > 0 {
			answer := ""
			for _, task := range t.tasks {
				if answer != "" {
					answer += "\n\n"
				}
				answer += strconv.Itoa(task.ID) + ". "
				answer += task.text + " "
				answer += "by @" + task.owner.UserName + "\n"
				if task.executer == nil {
					answer += "/assign_" + strconv.Itoa(task.ID)
				} else {
					answer += "assignee: "
					if task.executer.ID == user.ID {
						answer += "я\n/unassign_" + strconv.Itoa(task.ID) + " /resolve_" + strconv.Itoa(task.ID)
					} else {
						answer += "@" + task.executer.UserName
					}
				}

				t.bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					answer,
				))
			}
		} else {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Нет задач",
			))
		}
	}
	if strings.Contains(command, "/assign") {
		tID, err := strconv.Atoi(command[8:])
		if err != nil {
			return err
		}
		taskID := -1
		for i, t := range t.tasks {
			if t.ID == tID {
				taskID = i + 1
				break
			}
		}

		if taskID == -1 {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Задачи с id=`+command[8:]+` нет`,
			))
		}

		prev := t.tasks[taskID-1].executer
		reset := prev != nil
		t.tasks[taskID-1].executer = user

		t.bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			`Задача "`+t.tasks[taskID-1].text+`" назначена на вас`,
		))
		if reset {
			t.bot.Send(tgbotapi.NewMessage(
				int64(prev.ID),
				`Задача "`+t.tasks[taskID-1].text+`" назначена на @`+user.UserName,
			))

		} else {
			if t.tasks[taskID-1].owner.ID != t.tasks[taskID-1].executer.ID {
				t.bot.Send(tgbotapi.NewMessage(
					int64(t.tasks[taskID-1].owner.ID),
					`Задача "`+t.tasks[taskID-1].text+`" назначена на @`+user.UserName,
				))
			}
		}
	}
	if strings.Contains(command, "/unassign") {
		tID, err := strconv.Atoi(command[10:])
		if err != nil {
			return err
		}

		taskID := -1
		for _, t := range t.tasks {
			if t.ID == tID {
				taskID = t.ID
			}
		}

		if taskID == -1 {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Задачи с id=`+command[10:]+` нет`,
			))
		}

		if t.tasks[taskID-1].executer.ID != user.ID {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Задача не на вас`,
			))
		} else {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Принято`,
			))

			t.bot.Send(tgbotapi.NewMessage(
				int64(t.tasks[taskID-1].owner.ID),
				`Задача "`+t.tasks[taskID-1].text+`" осталась без исполнителя`,
			))
			t.tasks[taskID-1].executer = nil
		}
	}
	if strings.Contains(command, "/resolve") {
		tID, err := strconv.Atoi(command[9:])
		if err != nil {
			return err
		}

		taskID := -1
		for _, t := range t.tasks {
			if t.ID == tID {
				taskID = t.ID
			}
		}

		if taskID == -1 {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Задачи с id=`+command[10:]+` нет`,
			))
		}

		if t.tasks[taskID-1].executer.ID != user.ID {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Задача не на вас`,
			))
		} else {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Задача "`+t.tasks[taskID-1].text+`" выполнена`,
			))

			t.bot.Send(tgbotapi.NewMessage(
				int64(t.tasks[taskID-1].owner.ID),
				`Задача "`+t.tasks[taskID-1].text+`" выполнена @`+t.tasks[taskID-1].executer.UserName,
			))

			t.tasks = append(t.tasks[:taskID-1], t.tasks[taskID:]...)
		}
	}
	if strings.Contains(command, "/my") {
		if len(t.tasks) > 0 {
			answer := ""
			for _, task := range t.tasks {
				if task.executer != nil && task.executer.ID == user.ID {
					if answer != "" {
						answer += "\n\n"
					}
					answer += strconv.Itoa(task.ID) + ". "
					answer += task.text + " "
					answer += "by @" + task.owner.UserName + "\n"
					answer += "/unassign_" + strconv.Itoa(task.ID) + " /resolve_" + strconv.Itoa(task.ID)

					t.bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						answer,
					))
				}
			}
		} else {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Нет задач",
			))
		}
	}
	if strings.Contains(command, "/owner") {
		if len(t.tasks) > 0 {
			answer := ""
			for _, task := range t.tasks {
				if task.owner.ID == user.ID {
					if answer != "" {
						answer += "\n\n"
					}
					answer += strconv.Itoa(task.ID) + ". "
					answer += task.text + " "
					answer += "by @" + task.owner.UserName + "\n"
					if task.executer == nil {
						answer += "/assign_" + strconv.Itoa(task.ID)
					} else {
						answer += "assignee: "
						if task.executer.ID == user.ID {
							answer += "я\n/unassign_" + strconv.Itoa(task.ID) + " /resolve_" + strconv.Itoa(task.ID)
						} else {
							answer += "@" + task.executer.UserName
						}
					}

					t.bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						answer,
					))
				}
			}
		} else {
			t.bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Нет задач",
			))
		}
	}
	return nil
}

func startTaskBot(ctx context.Context) error {
	// сюда пишите ваш код
	t, err := newBot()
	if err != nil {
		return err
	}

	updates := t.bot.ListenForWebhook("/")

	go http.ListenAndServe(":8081", nil)
	fmt.Println("start listen :8081")

	for update := range updates {
		err = t.execUpdate(update)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
