package ui

import "fyne.io/fyne/v2"

const (
	NotificationSuccess = "Success"
	NotificationError   = "Error"
	NotificationInfo    = "Info"
)

func showNotification(title, content string) {
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   title,
		Content: content,
	})
}

func showErrorNotification(content string) {
	showNotification(NotificationError, content)
}

func showSuccessNotification(content string) {
	showNotification(NotificationSuccess, content)
}
