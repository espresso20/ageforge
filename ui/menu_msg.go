package ui

// menuModalPage is the page a dialog over the menu goes on.
const menuModalPage = "menu_modal"

// showMenuMsg shows msg in a dialog with an OK button.
func showMenuMsg(app *appLoop, pages *nativePages, msg string) {
	showButtonModal(app, pages, menuModalPage, msg, []string{txtOK.String()}, func(int, string) {})
}
