package resolve

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

const (
	// TODO: remove to baetyl-go and exposed by a method
	portsFile = "var/lib/baetyl/run/services.yml"
)

type Mapping struct {
	Ports map[string]PortsInfo `yaml:"ports,omitempty"`
}

type PortsInfo struct {
	Items  []int `yaml:"items,omitempty"`
	offset int
}

type NativeResolver struct {
	watcher *fsnotify.Watcher
	mapping *Mapping
	log     *log.Logger
	sync.RWMutex
}

func (i *PortsInfo) Next() (int, error) {
	if len(i.Items) == 0 {
		return 0, errors.New("ports of service are empty in ports mapping file")
	}
	port := i.Items[i.offset]
	i.offset++
	if i.offset == len(i.Items) {
		i.offset = 0
	}
	return port, nil
}

func NewNativeResolver(_ context.Context) (Resolver, error) {
	resolver := &NativeResolver{
		mapping: new(Mapping),
		log:     log.With(log.Any("resolve", "native")),
	}

	err := resolver.LoadMapping()
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					resolver.log.Warn("error when wait on the events channel")
					return
				}
				resolver.log.Debug("received a file event", log.Any("eventName", event.Name), log.Any("eventOp", event.Op))

				if event.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				resolver.log.Debug("load ports mapping file again", log.Error(err))
				err = resolver.LoadMapping()
				if err != nil {
					resolver.log.Warn("load ports mapping file failed", log.Error(err))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					resolver.log.Warn("error when wait on the events channel")
					return
				}
				resolver.log.Warn("error when wait on the events channel", log.Error(err))
			}
		}
	}()

	err = watcher.Add(portsFile)
	if err != nil {
		return nil, err
	}

	resolver.watcher = watcher
	return resolver, nil
}

func (n *NativeResolver) LoadMapping() error {
	n.Lock()
	defer n.Unlock()

	if !utils.FileExists(portsFile) {
		return errors.Errorf("ports mapping file (%s) doesn't exist", portsFile)
	}
	data, err := ioutil.ReadFile(portsFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, n.mapping)
	if err != nil {
		return err
	}
	return nil
}

func (n *NativeResolver) Resolve(service string) (address string, err error) {
	n.Lock()
	defer n.Unlock()

	portsInfo, ok := n.mapping.Ports[service]
	if !ok {
		return "", errors.New("no such service in ports mapping file")
	}
	port, err := portsInfo.Next()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("localhost:%d", port), nil
}

func (n *NativeResolver) Close() error {
	err := n.watcher.Close()
	if err != nil {
		n.log.Warn("failed to close file watcher", log.Error(err))
	}
	return nil
}
