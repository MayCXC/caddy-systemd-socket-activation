package networks

import (
	"context"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/coreos/go-systemd/v22/activation"
	"net"
	"os"
	"strconv"
	"strings"
)

var nameToFiles map[string][]*os.File

func init() {
	nameToFiles = filesWithNames()
	caddy.RegisterNetwork("sd", getListener)
	caddy.RegisterNetwork("sdgram", getListener)
	caddyhttp.RegisterNetworkHTTP3("sdgram", "sdgram")
}

// FilesWithNames maps fd names to a set of os.File pointers.
func filesWithNames() map[string][]*os.File {
	files := activation.Files(true)
	filesWithNames := map[string][]*os.File{}

	for _, f := range files {
		current, ok := filesWithNames[f.Name()]

		if !ok {
			current = []*os.File{}
			filesWithNames[f.Name()] = current
		}

		filesWithNames[f.Name()] = append(current, f)
	}

	return filesWithNames
}

func getListener(ctx context.Context, network, addr string, cfg net.ListenConfig) (any, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	li := strings.LastIndex(host, "/")
	if li < 0 {
		li = len(host)
	}
	name := host[:li]
	index := host[li+2:]

	i, err := strconv.ParseUint(index, 0, strconv.IntSize)
	if err != nil {
		return nil, err
	}

	files, ok := nameToFiles[name]
	if !ok {
		return nil, fmt.Errorf("invalid listen fd name: %s", name)
	}

	if uint64(len(files)) <= i {
		return nil, fmt.Errorf("invalid listen fd index: %d", i)
	}
	file := files[i]

	switch network {
	case "sd":
		return net.FileListener(file)
	case "sdgram":
		return net.FilePacketConn(file)
	}

	return nil, fmt.Errorf("invalid network: %s", network)
}
