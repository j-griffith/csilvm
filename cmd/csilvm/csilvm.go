package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/mesosphere/csilvm/pkg/csilvm"
)

const (
	defaultDefaultFs         = "xfs"
	defaultDefaultVolumeSize = 10 << 30
)

func main() {
	// Configure flags
	vgnameF := flag.String("volume-group", "", "The name of the volume group to manage")
	pvnamesF := flag.String("devices", "", "A comma-seperated list of devices in the volume group")
	defaultFsF := flag.String("default-fs", defaultDefaultFs, "The default filesystem to format new volumes with")
	defaultVolumeSizeF := flag.Uint64("default-volume-size", defaultDefaultVolumeSize, "The default volume size in bytes")
	socketFileF := flag.String("unix-addr", "", "The path to the listening unix socket file")
	removeF := flag.Bool("remove-volume-group", false, "If set, the volume group will be removed when ProbeNode is called.")
	profileF := flag.String("profile", "", "The volume group profile")
	flag.Parse()
	// Setup logging
	logprefix := fmt.Sprintf("[%s]", *vgnameF)
	logflags := log.LstdFlags | log.Lshortfile
	csilvm.SetLogger(log.New(os.Stderr, logprefix, logflags))
	// Setup socket listener
	lis, err := net.Listen("unix", *socketFileF)
	if err != nil {
		log.Fatalf("[%s] Failed to listen: %v", *vgnameF, err)
	}
	// Setup server
	grpcServer := grpc.NewServer()
	var opts []csilvm.ServerOpt
	opts = append(opts, csilvm.DefaultVolumeSize(*defaultVolumeSizeF))
	if *removeF {
		opts = append(opts, csilvm.RemoveVolumeGroup())
	}
	if *profileF != "" {
		opts = append(opts, csilvm.Profile(*profileF))
	}
	s := csilvm.NewServer(*vgnameF, strings.Split(*pvnamesF, ","), *defaultFsF, opts...)
	csi.RegisterIdentityServer(grpcServer, s)
	csi.RegisterControllerServer(grpcServer, s)
	csi.RegisterNodeServer(grpcServer, s)
	grpcServer.Serve(lis)
}

type vgLogger struct {
	vgname string
}

func (l vgLogger) Print(v ...interface{}) {
	log.Print(fmt.Sprintf("[%s]", l.vgname), v...)
}

func (l vgLogger) Printf(format string, v ...interface{}) {
	log.Print(fmt.Sprintf("[%s]", l.vgname)+" "+format, v...)
}