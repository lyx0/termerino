package main

// A simple example that shows how to send activity to Bubble Tea in real-time
// through a channel.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gempir/go-twitch-irc/v4"
)

// A message used to indicate that activity has occurred. In the real world (for
// example, chat) this would contain actual data.
type responseMsg struct {
	msg  string
	user string
}

// Simulate a process that sends events at an irregular interval in real time.
// In this case, we'll send events on the channel at a random interval between
// 100 to 1000 milliseconds. As a command, Bubble Tea will run this
// asynchronously.
func listenForActivity(sub chan responseMsg) tea.Cmd {
	// or client := twitch.NewAnonymousClient() for an anonymous user (no write capabilities)
	client := twitch.NewClient("justinfan123", "oauth:123123123")

	return func() tea.Msg {
		for {
			client.OnPrivateMessage(func(message twitch.PrivateMessage) {
				fmt.Println(message.Message)

				user := message.User.DisplayName
				msg := message.Message
				sub <- responseMsg{user, msg}
			})

			client.Join("nourylul")

			err := client.Connect()
			if err != nil {
				panic(err)
			}
			// user := "nourylul"
			// msg := "xD good message"
			// sub <- responseMsg{user, msg}
		}
	}
}

type (
	errMsg error
)

// A command that waits for the activity on a channel.
func waitForActivity(sub chan responseMsg) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

type model struct {
	sub     chan responseMsg // where we'll receive activity notifications
	spinner spinner.Model

	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		listenForActivity(m.sub), // generate activity
		waitForActivity(m.sub),   // wait for activity
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case responseMsg:
		m.messages = append(m.messages, m.senderStyle.Render(fmt.Sprintf("#%s: %s", msg.user, msg.msg)))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		return m, waitForActivity(m.sub) // wait for next event

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	switch msg.Type {
	// 	case tea.KeyCtrlC, tea.KeyEsc:
	// 		fmt.Println(m.textarea.Value())
	// 		return m, tea.Quit
	// 	case tea.KeyEnter:
	// 		m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
	// 		m.viewport.SetContent(strings.Join(m.messages, "\n"))
	// 		m.textarea.Reset()
	// 		m.viewport.GotoBottom()
	// 	}

	// // We handle errors just like any other message
	// case errMsg:
	// 	m.err = msg
	// 	return m, nil
	// }

	// return m, tea.Batch(tiCmd, vpCmd)
	// switch msg.(type) {
	// case tea.KeyMsg:
	// default:
	// 	return m, nil
	// }
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
