package kevdagoats_plugins

import (
	"context"
	"github.com/LeakIX/l9format"
	"github.com/jlaffaye/ftp"
	"log"
	"net"
	"time"
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
	// Instantiate the client (TODO: Need to consider weak creds)
	// Would like to use GetL9NetworkConnection here but that means we cannot specify a timeout
	client, err := ftp.DialTimeout(net.JoinHostPort(event.Ip, event.Port), time.Second*5)
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
	}
	// Lets get some stats, NOT using walk function due to massive time penalties
	files, err := client.List(".")
	if err != nil {
		return false
	}
	// Compute total size
	for _, file := range files {
		event.Leak.Dataset.Size += int64(file.Size)
	}
	event.Leak.Dataset.Files = int64(len(files))
	event.Service.Software.Name = "FTP"
	event.Leak.Severity = l9format.SEVERITY_INFO
	event.Summary = "FTP server is open and unprotected\n"
	event.Leak.Type = "open_database"
	return true
}
