package screen

import (
	"fmt"
	"log"
	"net"
	"guessh/cmd/cli/transport"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Execution error: %v", r)
			return
		}
	}()

	file, err := tea.LogToFile("cli.log", "")
	if err != nil {
		log.Fatalf("tea.LogToFile returned error: %v\n", err)
	}
	defer file.Close()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	conn      net.Conn
	sub       chan transport.EventMsg
	textInput textinput.Model
	events    string
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "{\"type\":\"make_guess\", ...}"
	ti.Focus()
	ti.CharLimit = 1024
	ti.Width = 150

	conn, err := net.Dial("tcp", "localhost:2480")
	if err != nil {
		log.Fatalf("net.Dial error: %v", err)
	}

	return model{
		conn:      conn,
		sub:       make(chan transport.EventMsg),
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		transport.ListenForActivity(m.conn, m.sub),
		transport.WaitForEvent(m.sub))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			transport.SendMessage(m.conn, m.textInput.Value())
			m.textInput.SetValue("")
		}
	case transport.EventMsg:
		m.events += fmt.Sprintf("%s\n", string(msg))
		return m, transport.WaitForEvent(m.sub)
	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"Received events: %s\nSend event: %s\n",
		m.events,
		m.textInput.View(),
	)
}
