package windows

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/k3s-io/k3s/tests/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Valid nodeOS: generic/ubuntu2004, opensuse/Leap-15.3.x86_64
var nodeOS = flag.String("nodeOS", "generic/ubuntu2004", "operating system for linux nodes")
var serverCount = flag.Int("serverCount", 1, "number of server nodes")
var windowsAgentCount = flag.Int("windowsAgentCount", 1, "number of windows agent nodes")

const defaultWindowsOS = "jborean93/WindowsServer2022"

func Test_WindowsValidation(t *testing.T) {
	flag.Parse()
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	RunSpecs(t, "Validate Windows Test Suite", suiteConfig, reporterConfig)
}

func createVMs(nodeOS string, serverCount, linuxAgentCount, windowsAgentCount int) ([]string, []string, []string, error) {
	serverNodeNames := []string{}
	for i := 0; i < serverCount; i++ {
		serverNodeNames = append(serverNodeNames, "server-"+strconv.Itoa(i))
	}
	windowsAgentNames := []string{}
	for i := 0; i < linuxAgentCount; i++ {
		windowsAgentNames = append(windowsAgentNames, "windows-agent-"+strconv.Itoa(i))
	}
	nodeRoles := strings.Join(serverNodeNames, " ") + " " + " " + strings.Join(windowsAgentNames, " ")
	nodeRoles = strings.TrimSpace(nodeRoles)
	nodeBoxes := strings.Repeat(nodeOS+" ", serverCount)
	nodeBoxes += strings.Repeat(defaultWindowsOS+" ", windowsAgentCount)
	nodeBoxes = strings.TrimSpace(nodeBoxes)

	cmd := fmt.Sprintf("NODE_ROLES=\"%s\" NODE_BOXES=\"%s\" %s vagrant up &> vagrant.log", nodeRoles, nodeBoxes, testOptions)
	fmt.Println(cmd)
	if _, err := e2e.RunCommand(cmd); err != nil {
		fmt.Println("Error Creating Cluster", err)
		return nil, nil, nil, err
	}
	return serverNodeNames, windowsAgentNames, nil
}

var (
	kubeConfigFile    string
	serverNodeNames   []string
	windowsAgentNames []string
)

var _ = ReportAfterEach(e2e.GenReport)
var _ = Describe("Verify Basic Cluster Creation", Ordered, func() {

	It("Starts up with no issues", func() {
		var err error
		serverNodeNames, windowsAgentNames, err = createVMs(*nodeOS, *serverCount, *windowsAgentCount)
		Expect(err).NotTo(HaveOccurred(), e2e.GetVagrantLog(err))
		fmt.Println("CLUSTER CONFIG")
		fmt.Println("OS:", *nodeOS)
		fmt.Println("Server Nodes:", serverNodeNames)
		fmt.Println("Windows Agent Nodes:", windowsAgentNames)
		kubeConfigFile, err = e2e.GenKubeConfigFile(serverNodeNames[0])
		Expect(err).NotTo(HaveOccurred())
	})

	It("Checks Node Status", func() {
		Eventually(func(g Gomega) {
			nodes, err := e2e.ParseNodes(kubeConfigFile, false)
			g.Expect(err).NotTo(HaveOccurred())
			for _, node := range nodes {
				g.Expect(node.Status).Should(Equal("Ready"))
			}
		}, "420s", "5s").Should(Succeed())
		_, err := e2e.ParseNodes(kubeConfigFile, true)
		Expect(err).NotTo(HaveOccurred())
	})

	//Copy flanneld.exe binary to windows VM

	//Create token on server node

	//Download flannel CNI binary

	//Write net-conf.json

})

var failed bool
var _ = AfterEach(func() {
	failed = failed || CurrentSpecReport().Failed()
})

var _ = AfterSuite(func() {
	if failed && !*ci {
		fmt.Println("FAILED!")
	} else {
		Expect(e2e.DestroyCluster()).To(Succeed())
		Expect(os.Remove(kubeConfigFile)).To(Succeed())
	}
})
