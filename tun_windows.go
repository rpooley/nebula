package nebula

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/songgao/water"
)

type Tun struct {
	Device string
	Cidr   *net.IPNet
	MTU    int
	InterfaceName string

	*water.Interface
}

func newTun(deviceName string, cidr *net.IPNet, defaultMTU int, routes []route, unsafeRoutes []route, txQueueLen int, interfaceName string) (ifce *Tun, err error) {
	if len(routes) > 0 {
		return nil, fmt.Errorf("Route MTU not supported in Windows")
	}
	if len(unsafeRoutes) > 0 {
		return nil, fmt.Errorf("unsafeRoutes not supported in Windows")
	}
	// NOTE: You cannot set the deviceName under Windows, so you must check tun.Device after calling .Activate()
	return &Tun{
		Cidr: cidr,
		MTU:  defaultMTU,
		InterfaceName: interfaceName,
	}, nil
}

func (c *Tun) Activate() error {
	var err error
	c.Interface, err = water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			ComponentID: "tap0901",
			Network:     c.Cidr.String(),
			InterfaceName: c.InterfaceName,
		},
	})
	if err != nil {
		return fmt.Errorf("Activate failed: %v", err)
	}

	c.Device = c.Interface.Name()

	// TODO use syscalls instead of exec.Command
	err = exec.Command(
		"netsh", "interface", "ipv4", "set", "address",
		fmt.Sprintf("name=%s", c.Device),
		"source=static",
		fmt.Sprintf("addr=%s", c.Cidr.IP),
		fmt.Sprintf("mask=%s", net.IP(c.Cidr.Mask)),
		"gateway=none",
	).Run()
	if err != nil {
		return fmt.Errorf("failed to run 'netsh' to set address: %s %s %s %s", err, c.Device,c.Cidr.IP,net.IP(c.Cidr.Mask))
	}
	err = exec.Command(
		"netsh", "interface", "ipv4", "set", "interface",
		c.Device,
		fmt.Sprintf("mtu=%d", c.MTU),
	).Run()
	if err != nil {
		return fmt.Errorf("failed to run 'netsh' to set MTU: %s", err)
	}

	return nil
}

func (c *Tun) WriteRaw(b []byte) error {
	_, err := c.Write(b)
	return err
}
