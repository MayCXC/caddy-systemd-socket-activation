package networks

import (
	"context"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/coreos/go-systemd/v22/activation"
	"net"
	"os"
	"path/filepath"
	"strconv"
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

	paths := filepath.SplitList(host)
	if len(paths) != 2 {
		return nil, fmt.Errorf("invalid listen address: %s/%s", network, addr)
	}

	name := paths[0]

	index, err := strconv.ParseUint(paths[1], 0, strconv.IntSize)
	if err != nil {
		return nil, err
	}

	files, ok := nameToFiles[name]
	if !ok {
		return nil, fmt.Errorf("invalid listen fd name: %s", name)
	}

	if uint64(len(files)) <= index {
		return nil, fmt.Errorf("invalid listen fd index: %d", index)
	}
	file := files[index]

	switch network {
	case "sd":
		return net.FileListener(file)
	case "sdgram":
		return net.FilePacketConn(file)
	}

	return nil, fmt.Errorf("invalid network: %s", network)
}
