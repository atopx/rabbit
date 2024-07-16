package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
)

const (
	XPathMatcher = "XPath"
	CSSSelector  = "CSS Selector"
	RegexMathcer = "Regex"
)

func main() {
	a := app.New()
	w := a.NewWindow("Developer")

	// 输入区域
	inputText := widget.NewMultiLineEntry()
	inputText.SetPlaceHolder("Input match text here...")
	inputText.SetMinRowsVisible(12) // 增加文本输入框的高度

	// 文件选择按钮
	fileButton := widget.NewButton("Select File", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				data, _ := io.ReadAll(reader)
				inputText.SetText(string(data))
			}
		}, w).Show()
	})

	// URL输入框
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL here...")

	// 获取URL内容按钮
	fetchURLButton := widget.NewButton("Fetch URL", func() {
		if _, err := url.Parse(urlEntry.Text); err != nil {
			inputText.SetText("Invalid internet url")
			return
		}
		resp, err := http.Get(urlEntry.Text)
		if err != nil {
			inputText.SetText(fmt.Sprintf("Fetch URL error: %v", err))
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		inputText.SetText(string(body))
	})

	// 匹配模式选择框
	matchMode := widget.NewSelect([]string{XPathMatcher, CSSSelector, RegexMathcer}, func(string) {})
	matchMode.SetSelected(XPathMatcher)

	// 匹配器代码输入框
	matcherEntry := widget.NewEntry()
	matcherEntry.SetPlaceHolder("Enter matcher code here...")

	// 结果显示区域
	resultLabel := widget.NewMultiLineEntry()
	resultLabel.SetPlaceHolder("Result will be displayed here...")
	resultLabel.SetMinRowsVisible(12) // 增加结果展示区域的高度
	resultLabel.Disable()             // 结果展示区域不可编辑

	// 实时匹配功能
	matcherEntry.OnChanged = func(code string) {
		text := inputText.Text
		mode := matchMode.Selected

		var result string
		var err error

		switch mode {
		case XPathMatcher:
			result, err = matchXPath(text, code)
		case CSSSelector:
			result, err = matchCSS(text, code)
		case RegexMathcer:
			result, err = matchRegex(text, code)
		}

		if err != nil {
			resultLabel.SetText(fmt.Sprintf("Error: %v", err))
		} else {
			resultLabel.SetText(result)
		}
	}

	// 布局
	content := container.NewVBox(
		container.NewHBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(590, 40)), urlEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), fetchURLButton),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 40)), fileButton),
		),
		inputText,
		container.NewGridWithColumns(3,
			matchMode,
			widget.NewButton("Copy Code", func() {
				w.Clipboard().SetContent(matcherEntry.Text)
			}),
			widget.NewButton("Copy Result", func() {
				w.Clipboard().SetContent(resultLabel.Text)
			}),
		),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(800, 40)), matcherEntry),
		resultLabel,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("Matcher Tool", content),
	)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(800, 670))
	w.SetFixedSize(true)
	w.SetFullScreen(false)
	w.ShowAndRun()
}

func matchXPath(input, code string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		return "", err
	}
	expr, err := xpath.Compile(code)
	if err != nil {
		return "", err
	}
	result := htmlquery.QuerySelectorAll(doc.Nodes[0], expr)
	var output []string
	for _, node := range result {
		output = append(output, htmlquery.InnerText(node))
	}
	return strings.Join(output, "\n"), nil
}

func matchCSS(input, code string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		return "", err
	}
	result := doc.Find(code)
	var output []string
	result.Each(func(i int, s *goquery.Selection) {
		output = append(output, s.Text())
	})
	return strings.Join(output, "\n"), nil
}

func matchRegex(input, code string) (string, error) {
	re, err := regexp.Compile(code)
	if err != nil {
		return "", err
	}
	result := re.FindAllString(input, -1)
	return strings.Join(result, "\n"), nil
}
