package controllers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yudolevich/kube-dev-course/example/operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TBot struct {
	*tgbotapi.BotAPI
	kube client.Client
	uid  int64
}

func NewTBot(token string, uid int64, client client.Client) (*TBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TBot{BotAPI: bot, uid: uid, kube: client}, nil
}

func (b *TBot) Start(ctx context.Context) error {
	for update := range b.GetUpdatesChan(tgbotapi.NewUpdate(0)) {
		if update.CallbackQuery == nil {
			continue
		}

		name := strings.Split(update.CallbackQuery.Data, "/")
		if len(name) != 2 {
			continue
		}

		b.kube.Status().Patch(ctx, &v1alpha1.Nginx{
			ObjectMeta: v1.ObjectMeta{Name: name[1], Namespace: name[0]},
		}, client.RawPatch(types.MergePatchType, []byte(`{"status":{"approved": true}}`)),
		)
	}

	return nil
}

func (b *TBot) SendDeploy(nginx *v1alpha1.Nginx) {
	msg := tgbotapi.NewMessage(
		b.uid,
		fmt.Sprintf(`Nginx: %s/%s
		replicas: %d
		index.html: %s`,
			nginx.GetNamespace(), nginx.GetName(),
			nginx.Spec.Replicas, nginx.Spec.Index),
	)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"deploy",
				fmt.Sprintf("%s/%s", nginx.GetNamespace(), nginx.GetName()),
			),
		),
	)

	b.Send(msg)
}
