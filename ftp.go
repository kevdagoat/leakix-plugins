package kevdagoats_plugins

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/LeakIX/l9format"
	"github.com/jlaffaye/ftp"
)

type FTPOpenPlugin struct {
	l9format.ServicePluginBase
}

func (FTPOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (FTPOpenPlugin) GetProtocols() []string {
	return []string{"ftp"}
}

func (FTPOpenPlugin) GetName() string {
	return "FTPOpenPlugin"
}

func (FTPOpenPlugin) GetStage() string {
	return "open"
}

func (plugin FTPOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) bool {
	// Get deadline
	deadline, _ := ctx.Deadline()
	// Instantiate the client (TODO: Need to consider weak creds)
	client, err := ftp.Dial(
		net.JoinHostPort(event.Ip, event.Port),
		ftp.DialWithShutTimeout(time.Until(deadline)),
		ftp.DialWithDialFunc(
			func(network, address string) (net.Conn, error) {
				return plugin.DialContext(ctx, network, address)
			}))
	if err != nil {
		// Not a FTP server
		log.Print("FTP connect failed, most likely not a FTP server: ", err)
		return false
	}
	defer client.Quit()
	// Login with anonymous creds
	err = client.Login("anonymous", "anonymous")
	if err != nil {
		// Creds invalid
		log.Print("FTP login failed: ", err)
		return false
	}
	// Lets get some stats, NOT using walk function due to massive time penalties
	files, err := client.List(".")
	if err != nil {
		return false
	}
	// Get current directory
	cwd, err := client.CurrentDir()
	if err != nil {
		// PWD not supported?
		log.Print("FTP PWD command failed: ", err)
		return false
	}
	event.Summary = fmt.Sprintf("Found %d files in %s :\n", len(files), cwd)
	// Compute total size
	for _, file := range files {
		event.Summary += fmt.Sprintf("%s	%d bytes\n", file.Name, file.Size)
		event.Leak.Dataset.Size += int64(file.Size)
	}
	event.Leak.Dataset.Files = int64(len(files))
	event.Service.Software.Name = "FTP"
	event.Leak.Severity = l9format.SEVERITY_INFO
	event.Summary = "FTP server is open and unprotected\n"
	event.Leak.Type = "open_database"
	return true
}
