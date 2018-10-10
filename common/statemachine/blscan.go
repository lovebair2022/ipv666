package statemachine

import (
	"github.com/lavalamp-/ipv666/common/config"
	"github.com/lavalamp-/ipv666/common/data"
	"log"
	"time"
	"github.com/lavalamp-/ipv666/common/shell"
	"github.com/rcrowley/go-metrics"
	"github.com/lavalamp-/ipv666/common/fs"
)

var blCandidateCounter = metrics.NewCounter()
var blCandidateResponseCounter = metrics.NewCounter()
var zmapNetsDurationTimer = metrics.NewTimer()

func init() {
	metrics.Register("zmap_nets_scan_duration", zmapNetsDurationTimer)
	metrics.Register("zmap_nets_cand_counter", blCandidateCounter)
	metrics.Register("zmap_nets_cand_response_counter", blCandidateResponseCounter)
}

func zmapScanNetworkRanges(conf *config.Configuration) (error) {
	addrsPath, err := data.GetMostRecentFilePathFromDir(conf.GetNetworkScanTargetsDirPath())
	if err != nil {
		return err
	}
	addrsCount, err := fs.CountLinesInFile(addrsPath)
	if err != nil {
		log.Printf("Could not read lines in file '%s': %e", addrsPath, err)
		return err
	}
	blCandidateCounter.Inc(int64(addrsCount))
	log.Printf("Going to scan blacklist candidate addresses in file at path '%s' (%d addresses).", addrsPath, addrsCount)
	outputPath := getTimedFilePath(conf.GetNetworkScanResultsDirPath())
	log.Printf("Results will be written to file '%s'.", outputPath)
	start := time.Now()
	_, err = shell.ZmapScanFromConfig(conf, addrsPath, outputPath)
	elapsed := time.Since(start)
	zmapNetsDurationTimer.Update(elapsed)
	log.Printf("Zmap scan took approximately %s.", elapsed)
	if err != nil {
		log.Printf("An error was thrown when trying to run zmap: %s", err)
		return err
	}
	resultsCount, err := fs.CountLinesInFile(outputPath)
	if err != nil {
		log.Printf("Could not read lines in file '%s': %e", outputPath, err)
		return err
	}
	blCandidateResponseCounter.Inc(int64(resultsCount))
	log.Printf("%d addresses responded to the ping scan.", resultsCount)
	return nil
}
//
//func zmapScanNetworkRanges(conf *config.Configuration) (error) {
//
//	// Find the target network groups file
//	netsPath, err := data.GetMostRecentFilePathFromDir(conf.GetNetworkGroupDirPath())
//	if err != nil {
//		return err
//	}
//
//	// Load the network groups
//	log.Printf("Loading network groups")
//	nets, err := addressing.GetAddressListFromHexStringsFile(netsPath)
//	if err != nil {
//		return err
//	}
//
//	start := time.Now()
//	// Generate random addressing in each network
//	log.Printf("Generating %d addressing in each network range", conf.NetworkPingCount)
//	rand.Seed(time.Now().UTC().UnixNano())
//	file, err := ioutil.TempFile("/tmp", "addrs")
//	if err != nil {
//		return err
//	}
//	var netRanges [][]addressing.IPv6Address
//	for _, net := range(nets.Addresses) {
//		var netRange []addressing.IPv6Address
//		for x := 0; x < conf.NetworkPingCount; x++ {
//			addr := addressing.IPv6Address{net.Content}
//			for x := conf.NetworkGroupingSize; x < 128; x++ {
//				byteOff := (int)(x/8)
//				bitOff := (uint)(x-(byteOff*8))
//				byteMask := (byte)(^(rand.Intn(2) << bitOff))
//				addr.Content[byteOff] |= (byte)(^byteMask)
//			}
//			netRange = append(netRange, addr)
//			file.WriteString(fmt.Sprintf("%s\n", addr.String()))
//		}
//		netRanges = append(netRanges, netRange)
//	}
//	file.Close()
//	elapsed := time.Since(start)
//	addrNetsGenerationTimer.Update(elapsed)
//
//	// Scan the addressing
//	inputPath, err := filepath.Abs(file.Name())
//	if err != nil {
//		return err
//	}
//	file, err = ioutil.TempFile("/tmp", "addrs-scanned")
//	if err != nil {
//		return err
//	}
//	outputPath, err := filepath.Abs(file.Name())
//	if err != nil {
//		return err
//	}
//	log.Printf(
//		"Now Zmap scanning IPv6 addressing found in file at path '%s'. Results will be written to '%s'.",
//		inputPath,
//		outputPath,
//	)
//	start = time.Now()
//	_, err = shell.ZmapScanFromConfig(conf, inputPath, outputPath)
//	elapsed = time.Since(start)
//	if err != nil {
//		log.Printf("An error was thrown when trying to run zmap: %s", err)
//		log.Printf("Zmap elapsed time was %s.", elapsed)
//		return err
//	}
//	zmapNetsDurationTimer.Update(elapsed)
//	liveCount, err := fs.CountLinesInFile(outputPath)
//	if err != nil {
//		log.Printf("Error when counting lines in file '%s': %e", outputPath, err)
//		if conf.ExitOnFailedMetrics {
//			return err
//		}
//	}
//	liveAddrNetsGauge.Update(int64(liveCount))
//	log.Printf("Zmap completed successfully in %s. Results written to file at '%s'.", elapsed, outputPath)
//
//	// Blacklist networks with 100% response rate
//	blacklistPath := getTimedFilePath(conf.GetNetworkBlacklistDirPath())
//	log.Printf("Writing network blacklist to %s.", blacklistPath)
//	file, err = os.OpenFile(blacklistPath, os.O_WRONLY|os.O_CREATE, 0600)
//	if err != nil {
//		return err
//	}
//	addrs, err := addressing.GetAddressListFromHexStringsFile(outputPath)
//	if err != nil {
//		return err
//	}
//	for pos, netRange := range netRanges {
//		addrMiss := false
//		for _, netAddr := range netRange {
//			found := false
//			for _, addr := range addrs.Addresses {
//				if netAddr.Content == addr.Content {
//					found = true
//					break
//				}
//			}
//			if found == false {
//				addrMiss = true
//				break
//			}
//		}
//
//		// If there were no response misses blacklist this network range
//		if addrMiss == false {
//			file.WriteString(fmt.Sprintf("%s\n", nets.Addresses[pos].String()))
//		}
//	}
//	file.Close()
//
//	return nil
//}
