package daemons

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/swarleynunez/hidra/core/managers"
	"github.com/swarleynunez/hidra/core/types"
	"github.com/swarleynunez/hidra/core/utils"
	"github.com/swarleynunez/hidra/inputs"
	"math/rand"
)

// MonitorV1 //
func checkStateRules(ctx context.Context, rccs map[string]types.CycleCounter, minter, ctime uint64, ccache map[uint64]bool) {

	state := managers.GetState()

	for _, rule := range inputs.Rules {
		// Current spec value (variable for different value types)
		var usage interface{}

		switch rule.Resource {
		case types.CpuResource:
			usage = selectCpuMetric(rule.MetricType, state)
		case types.MemResource:
			usage = selectMemMetric(rule.MetricType, state)
		case types.DiskResource:
			usage = selectDiskMetric(rule.MetricType, state)
		case types.PktSentResource:
			usage = selectPktSentMetric(rule.MetricType, state)
		case types.PktRecvResource:
			usage = selectPktRecvMetric(rule.MetricType, state)
		default:
			utils.CheckError(errUnknownSpec, utils.WarningMode)
			continue // Drop rule check
		}

		if usage == nil {
			utils.CheckError(errBoundNotImplemented, utils.WarningMode)
			continue // Drop rule check
		}

		// Get the rcc
		rcc := rccs[rule.NameID]

		// Count a measure if the rcc has already started
		if rcc.Measures > 0 {
			rcc.Measures++
		}

		// Rule checking
		if ok, err := utils.CompareValues(usage, rule.Comparator, rule.Limit); ok {

			// Start the rcc
			if rcc.Measures == 0 {
				rcc.Measures++
			}

			// Count a trigger
			rcc.Triggers++

			if rcc.Measures == ctime/minter && rcc.Measures == rcc.Triggers {
				runRuleAction(ctx, &rule, ccache, usage)
			}
		} else {
			utils.CheckError(err, utils.WarningMode)
		}

		// Reset the rcc
		if rcc.Measures == ctime/minter {
			//rcc = cycle{}
		}

		// Update the rcc for the next state checking
		rccs[rule.NameID] = rcc
	}
}

// MonitorV2 //
func monitorNetwork(iface, nodePort string, lossProb, maxLatency uint64, nodeStore types.NodeStore, pktCounter *types.PacketCounter) {

	// Open interface
	handle, err := pcap.OpenLive(iface, 65536, false, pcap.BlockForever)
	utils.CheckError(err, utils.FatalMode)
	defer handle.Close()

	// TODO. Filtering by UDP and ports (due to the emulation of fog nodes)
	err = handle.SetBPFFilter("udp and port " + nodePort + " and !port 30301")
	utils.CheckError(err, utils.FatalMode)

	// Use the handle as a packet source to process all packets
	pktSrc := gopacket.NewPacketSource(handle, handle.LinkType())
	for pkt := range pktSrc.Packets() {
		// Counting all packets
		pktCounter.Total++

		processPacket(pkt, nodePort, lossProb, maxLatency, nodeStore, pktCounter)

		// Bounding the experiment
		/*if pktCounter.Total == pktCounter.Max {
			managers.PrintFinalStatistics(nodeStore, pktCounter)
			os.Exit(0)
		}*/
	}
}

func processPacket(pkt gopacket.Packet, nodePort string, lossProb, maxLatency uint64, nodeStore types.NodeStore, pktCounter *types.PacketCounter) {

	// Get packet network/transport info
	srcIP := pkt.NetworkLayer().NetworkFlow().Src().String()
	srcPort := pkt.TransportLayer().TransportFlow().Src().String()
	dstIP := pkt.NetworkLayer().NetworkFlow().Dst().String()
	dstPort := pkt.TransportLayer().TransportFlow().Dst().String()

	// Packet performance simulation
	lost, latency := simulatePacketPerformance(lossProb, maxLatency)

	// Packet sender/receiver?
	var nodeAddr common.Address
	if srcPort == nodePort { // Outgoing packets
		if !lost {
			pktCounter.Sent++

			// Debug
			//fmt.Print("[", time.Now().UnixMilli(), "] ", "Packet ", pktCounter.Total,
			//	" sent to ", dstIP, ":", dstPort, " with ", latency, "ms latency\n")
		} else {
			// Debug
			//fmt.Print("[", time.Now().UnixMilli(), "] ", "Dropping packet...\n")
		}

		// Get the other fog node of the edge
		nodeAddr = managers.GetNodeAddressFromIP(dstIP, dstPort)
	} else { // Incoming packets
		if !lost {
			pktCounter.Recv++

			// Debug
			//fmt.Print("[", time.Now().UnixMilli(), "] ", "Packet ", pktCounter.Total,
			//	" received from ", srcIP, ":", srcPort, " with ", latency, "ms latency\n")
		} else {
			// Debug
			//fmt.Print("[", time.Now().UnixMilli(), "] ", "Dropping packet...\n")
		}

		// Get the other fog node of the edge
		nodeAddr = managers.GetNodeAddressFromIP(srcIP, srcPort)
	}

	if !utils.EmptyEthAddress(nodeAddr.String()) {
		// New fog nodes/peers
		if nodeStore[nodeAddr] == nil {
			nodeStore[nodeAddr] = &types.NodeInfo{}
		}

		// Update node's current epoch info
		nodeStore[nodeAddr].CurrentEpoch.TotalPackets++
		if !lost {
			nodeStore[nodeAddr].CurrentEpoch.OKPackets++
			nodeStore[nodeAddr].CurrentEpoch.Latencies = append(nodeStore[nodeAddr].CurrentEpoch.Latencies, latency)
		}
	}

	return
}

func simulatePacketPerformance(lossProb, maxLatency uint64) (lost bool, latency uint64) {

	if uint64(rand.Intn(100+1)) < lossProb {
		lost = true
	} else {
		latency = uint64(rand.Intn(int(maxLatency)) + 1)
	}

	return
}
