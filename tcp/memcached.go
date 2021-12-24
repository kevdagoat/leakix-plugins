package tcp

import (
	"context"
	"github.com/LeakIX/l9format"
	"github.com/memcachier/mc/v3"
	"log"
	"net"
	"strconv"
)

type MemcachedOpenPlugin struct {
	l9format.ServicePluginBase
}

func (MemcachedOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (MemcachedOpenPlugin) GetProtocols() []string {
	return []string{"memcached"}
}

func (MemcachedOpenPlugin) GetName() string {
	return "MemcachedOpenPlugin"
}

func (MemcachedOpenPlugin) GetStage() string {
	return "open"
}

func (plugin MemcachedOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) bool {
	// Instantiate the client (TODO: Need to consider weak creds)
	client := mc.NewMC(net.JoinHostPort(event.Ip, event.Port), "", "")
	defer client.Quit()
	// ping
	if err := client.NoOp(); err != nil {
		// Looks like we don't have a connection
		log.Print("Memcached server exited on ping (noop): ", err)
		return false
	}
	// get stats
	stats, err := client.Stats()
	if err != nil {
		log.Print("Memcached server aborted stats command: ", err)
		return false
	}
	serverStats := stats[net.JoinHostPort(event.Ip, event.Port)]
	event.Service.Software.Name = "Memcached"
	event.Service.Software.Version = serverStats["version"]
	event.Leak.Severity = l9format.SEVERITY_MEDIUM
	event.Summary = "Memcached is open and unprotected\n"
	event.Leak.Type = "open_database"
	// Type cast to integers
	if rows, err := strconv.Atoi(serverStats["total_items"]); err == nil {
		event.Leak.Dataset.Rows = int64(rows)
	}
	if rows, err := strconv.Atoi(serverStats["bytes"]); err == nil {
		event.Leak.Dataset.Size = int64(rows)
	}
	return true
}
