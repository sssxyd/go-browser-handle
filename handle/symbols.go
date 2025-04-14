package handle

import (
	"reflect"

	"github.com/playwright-community/playwright-go"
)

// Exported Symbols variable
var Symbols = map[string]reflect.Value{
	//playwright-go相关类型和方法
	"Locator": reflect.ValueOf((*playwright.Locator)(nil)), // Export Locator interface pointer type
	"Page":    reflect.ValueOf((*playwright.Page)(nil)),    // Export Page interface pointer type

	// 浏览器相关类型和方法
	"TabPage": reflect.ValueOf((*TabPage)(nil)), // Export TabPage interface pointer type
	"Browser": reflect.ValueOf((*Browser)(nil)), // Export Browser interface pointer type

	// Edge浏览器初始化方法
	"Edge":                          reflect.ValueOf(Edge),                          // Export Edge function
	"(*EdgeBrowserInstance).Listen": reflect.ValueOf((*EdgeBrowserInstance).Listen), // Export Listen method

	// TabPage的方法
	"(*TabPage).ID":               reflect.ValueOf((*TabPage)(nil)).MethodByName("ID"),
	"(*TabPage).Title":            reflect.ValueOf((*TabPage)(nil)).MethodByName("Title"),
	"(*TabPage).URL":              reflect.ValueOf((*TabPage)(nil)).MethodByName("URL"),
	"(*TabPage).Domain":           reflect.ValueOf((*TabPage)(nil)).MethodByName("Domain"),
	"(*TabPage).Close":            reflect.ValueOf((*TabPage)(nil)).MethodByName("Close"),
	"(*TabPage).IsClosed":         reflect.ValueOf((*TabPage)(nil)).MethodByName("IsClosed"),
	"(*TabPage).BringToFront":     reflect.ValueOf((*TabPage)(nil)).MethodByName("BringToFront"),
	"(*TabPage).OpenInNewTab":     reflect.ValueOf((*TabPage)(nil)).MethodByName("OpenInNewTab"),
	"(*TabPage).WaitSelector":     reflect.ValueOf((*TabPage)(nil)).MethodByName("WaitSelector"),
	"(*TabPage).QuerySelector":    reflect.ValueOf((*TabPage)(nil)).MethodByName("QuerySelector"),
	"(*TabPage).QuerySelectorAll": reflect.ValueOf((*TabPage)(nil)).MethodByName("QuerySelectorAll"),
	"(*TabPage).ClearLocalData":   reflect.ValueOf((*TabPage)(nil)).MethodByName("ClearLocalData"),
	"(*TabPage).Goto":             reflect.ValueOf((*TabPage)(nil)).MethodByName("Goto"),
	"(*TabPage).Evaluate":         reflect.ValueOf((*TabPage)(nil)).MethodByName("Evaluate"),
	"(*TabPage).Page":             reflect.ValueOf((*TabPage)(nil)).MethodByName("Page"),
	"(*TabPage).Reload":           reflect.ValueOf((*TabPage)(nil)).MethodByName("Reload"),
	"(*TabPage).GetCookies":       reflect.ValueOf((*TabPage)(nil)).MethodByName("GetCookies"),
	"(*TabPage).ApplyCookies":     reflect.ValueOf((*TabPage)(nil)).MethodByName("ApplyCookies"),

	// Browser的方法
	"(*Browser).Name":            reflect.ValueOf((*Browser)(nil)).MethodByName("Name"),
	"(*Browser).Port":            reflect.ValueOf((*Browser)(nil)).MethodByName("Port"),
	"(*Browser).TabPages":        reflect.ValueOf((*Browser)(nil)).MethodByName("TabPages"),
	"(*Browser).NewTabPage":      reflect.ValueOf((*Browser)(nil)).MethodByName("NewTabPage"),
	"(*Browser).DefaultPage":     reflect.ValueOf((*Browser)(nil)).MethodByName("DefaultPage"),
	"(*Browser).FindTabPage":     reflect.ValueOf((*Browser)(nil)).MethodByName("FindTabPage"),
	"(*Browser).SwitchToTabPage": reflect.ValueOf((*Browser)(nil)).MethodByName("SwitchToTabPage"),
	"(*Browser).CloseTabPage":    reflect.ValueOf((*Browser)(nil)).MethodByName("CloseTabPage"),
	"(*Browser).IsAlive":         reflect.ValueOf((*Browser)(nil)).MethodByName("IsAlive"),
	"(*Browser).Close":           reflect.ValueOf((*Browser)(nil)).MethodByName("Close"),
}
