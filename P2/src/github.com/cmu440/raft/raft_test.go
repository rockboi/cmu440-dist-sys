package raft

import "github.com/cmu440/rpc"
import "log"
import "sync"
import "testing"
import "runtime"
import crand "crypto/rand"
import "encoding/base64"
import "sync/atomic"
import "time"
import "fmt"

//
// Raft tests.
//
// We will use the original raft_test.go to test your code for grading.
// So, while you can modify this code to help you debug, please
// test with the original before submitting.
//

// The tester generously allows solutions to complete elections in one second
// (much more than the paper's range of timeouts).
const RaftElectionTimeout = 1000 * time.Millisecond

func TestInitialElection2A(t *testing.T) {
	fmt.Printf("==================== 3 SERVERS ====================\n")
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test (2A): Initial election\n")

	// is a leader elected?
	fmt.Printf("Checking current leader\n")
	cfg.checkOneLeader()

	fmt.Printf("======================= END =======================\n\n")
}

func TestReElection2A(t *testing.T) {
	fmt.Printf("==================== 3 SERVERS ====================\n")
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test (2A): Re-election\n")
	fmt.Printf("Basic 1 leader\n")
	leader1 := cfg.checkOneLeader()

	// if the leader disconnects, a new one should be elected.
	fmt.Printf("Disconnecting leader\n")
	cfg.disconnect(leader1)

	// a new leader should be elected
	fmt.Printf("Checking for a new leader\n")
	cfg.checkOneLeader()

	fmt.Printf("======================= END =======================\n\n")
}

func TestBasicAgree2B(t *testing.T) {
	fmt.Printf("==================== 5 SERVERS ====================\n")
	servers := 5
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test (2B): basic agreement\n")

	iters := 3
	for index := 1; index < iters+1; index++ {
		nd, _ := cfg.nCommitted(index)
		if nd > 0 {
			t.Fatalf("Some have committed before Start()")
		}

		xindex := cfg.one(index*100, servers)
		if xindex != index {
			t.Fatalf("Got index %v but expected %v", xindex, index)
		}
	}

	fmt.Printf("======================= END =======================\n\n")
}

func TestFailAgree2B(t *testing.T) {
	fmt.Printf("==================== 3 SERVERS ====================\n")
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test (2B): agreement despite \nfollower disconnection\n")

	cfg.one(101, servers)

	// follower network disconnection
	fmt.Printf("Checking one leader\n")
	leader := cfg.checkOneLeader()
	cfg.disconnect((leader + 1) % servers)

	fmt.Printf("Checking agreement with one disconnected peer\n")
	// agree despite two disconnected servers?
	cfg.one(102, servers-1)
	cfg.one(103, servers-1)
	time.Sleep(RaftElectionTimeout)
	cfg.one(104, servers-1)
	cfg.one(105, servers-1)

	// re-connect
	cfg.connect((leader + 1) % servers)
	fmt.Printf("Checking with one reconnected server\n")
	// agree with full set of servers?
	cfg.one(106, servers)
	time.Sleep(RaftElectionTimeout)
	cfg.one(107, servers)

	fmt.Printf("======================= END =======================\n\n")
}

func TestFailNoAgree2B(t *testing.T) {
	fmt.Printf("==================== 5 SERVERS ====================\n")
	servers := 5
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test (2B): no agreement without majority\n")

	fmt.Printf("Checking agreement\n")
	cfg.one(10, servers)

	// 3 of 5 followers disconnect
	leader := cfg.checkOneLeader()
	cfg.disconnect((leader + 1) % servers)
	cfg.disconnect((leader + 2) % servers)
	cfg.disconnect((leader + 3) % servers)

	fmt.Printf("Disconnected 3 out of 5 peers\n")

	index, _, ok := cfg.rafts[leader].Start(20)
	if !ok {
		t.Fatalf("Leader rejected Start()")
	}
	if index != 2 {
		t.Fatalf("Expected index 2, got %v", index)
	}

	time.Sleep(2 * RaftElectionTimeout)

	n, _ := cfg.nCommitted(index)
	if n > 0 {
		t.Fatalf("%v committed but no majority", n)
	}

	// repair
	cfg.connect((leader + 1) % servers)
	cfg.connect((leader + 2) % servers)
	cfg.connect((leader + 3) % servers)

	fmt.Printf("Reconnected all peers\n")

	// the disconnected majority may have chosen a leader from
	// among their own ranks, forgetting index 2.
	// or perhaps
	leader2 := cfg.checkOneLeader()
	index2, _, ok2 := cfg.rafts[leader2].Start(30)
	if !ok2 {
		t.Fatalf("Leader2 rejected Start()")
	}
	if index2 < 2 || index2 > 3 {
		t.Fatalf("Unexpected index %v", index2)
	}

	fmt.Printf("Checking agreement\n")
	cfg.one(1000, servers)

	fmt.Printf("======================= END =======================\n\n")
}

func TestConcurrentStarts2B(t *testing.T) {
	fmt.Printf("==================== 3 SERVERS ====================\n")
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test (2B): concurrent Start()s\n")

	var success bool
loop:
	for try := 0; try < 5; try++ {
		if try > 0 {
			// give solution some time to settle
			time.Sleep(3 * time.Second)
		}

		leader := cfg.checkOneLeader()
		_, term, ok := cfg.rafts[leader].Start(1)
		if !ok {
			// leader moved on really quickly
			continue
		}

		iters := 5
		var wg sync.WaitGroup
		is := make(chan int, iters)
		for ii := 0; ii < iters; ii++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				i, term1, ok := cfg.rafts[leader].Start(100 + i)
				if term1 != term {
					return
				}
				if !ok {
					return
				}
				is <- i
			}(ii)
		}

		wg.Wait()
		close(is)

		for j := 0; j < servers; j++ {
			if _, t, _ := cfg.rafts[j].GetState(); t != term {
				// term changed -- can't expect low RPC counts
				continue loop
			}
		}

		failed := false
		cmds := []int{}
		for index := range is {
			cmd := cfg.wait(index, servers, term)
			if ix, ok := cmd.(int); ok {
				if ix == -1 {
					// peers have moved on to later terms
					// so we can't expect all Start()s to
					// have succeeded
					failed = true
					break
				}
				cmds = append(cmds, ix)
			} else {
				t.Fatalf("Value %v is not an int", cmd)
			}
		}

		if failed {
			// avoid leaking goroutines
			go func() {
				for range is {
				}
			}()
			continue
		}

		for ii := 0; ii < iters; ii++ {
			x := 100 + ii
			ok := false
			for j := 0; j < len(cmds); j++ {
				if cmds[j] == x {
					ok = true
				}
			}
			if !ok {
				t.Fatalf("Cmd %v missing in %v", x, cmds)
			}
		}

		success = true
		break
	}

	if !success {
		t.Fatalf("Term changed too often")
	}

	fmt.Printf("======================= END =======================\n\n")
}

//
// Support for Raft tester
//
// We will use the original test file to test your code for grading
// so, while you can modify this code to help you debug, please
// test with the original before submitting.
//

func randstring(n int) string {
	b := make([]byte, 2*n)
	crand.Read(b)
	s := base64.URLEncoding.EncodeToString(b)
	return s[0:n]
}

type config struct {
	mu        sync.Mutex
	t         *testing.T
	net       *rpc.Network
	n         int
	done      int32 // tell internal threads to die
	rafts     []*Raft
	applyErr  []string      // from apply channel readers
	connected []bool        // whether each server is on the net
	endnames  [][]string    // the port file names each sends to
	logs      []map[int]int // copy of each server's committed entries
}

var ncpu_once sync.Once

func make_config(t *testing.T, n int, unreliable bool) *config {
	ncpu_once.Do(func() {
		if runtime.NumCPU() < 2 {
			fmt.Printf("warning: only one CPU, which may conceal locking bugs\n")
		}
	})
	runtime.GOMAXPROCS(4)
	cfg := &config{}
	cfg.t = t
	cfg.net = rpc.MakeNetwork()
	cfg.n = n
	cfg.applyErr = make([]string, cfg.n)
	cfg.rafts = make([]*Raft, cfg.n)
	cfg.connected = make([]bool, cfg.n)
	cfg.endnames = make([][]string, cfg.n)
	cfg.logs = make([]map[int]int, cfg.n)

	cfg.setunreliable(unreliable)

	cfg.net.LongDelays(true)

	// create a full set of Rafts.
	for i := 0; i < cfg.n; i++ {
		cfg.logs[i] = map[int]int{}
		cfg.start1(i)
	}

	// connect everyone
	for i := 0; i < cfg.n; i++ {
		cfg.connect(i)
	}

	return cfg
}

// shut down a Raft server.
func (cfg *config) crash1(i int) {
	cfg.disconnect(i)
	cfg.net.DeleteServer(i) // disable client connections to the server.

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	rf := cfg.rafts[i]
	if rf != nil {
		cfg.mu.Unlock()
		rf.Kill()
		cfg.mu.Lock()
		cfg.rafts[i] = nil
	}
}

//
// start or re-start a Raft.
// if one already exists, "kill" it first.
// allocate new outgoing port file names
// to isolate previous instance of
// this server. since we cannot really kill it.
//
func (cfg *config) start1(i int) {
	cfg.crash1(i)

	// a fresh set of outgoing ClientEnd names.
	// so that old crashed instance's ClientEnds can't send.
	cfg.endnames[i] = make([]string, cfg.n)
	for j := 0; j < cfg.n; j++ {
		cfg.endnames[i][j] = randstring(20)
	}

	// a fresh set of ClientEnds.
	ends := make([]*rpc.ClientEnd, cfg.n)
	for j := 0; j < cfg.n; j++ {
		ends[j] = cfg.net.MakeEnd(cfg.endnames[i][j])
		cfg.net.Connect(cfg.endnames[i][j], j)
	}

	cfg.mu.Lock()

	cfg.mu.Unlock()

	// listen to messages from Raft indicating newly committed messages.
	applyCh := make(chan ApplyMsg)
	go func() {
		for m := range applyCh {
			err_msg := ""
			if v, ok := (m.Command).(int); ok {
				cfg.mu.Lock()
				for j := 0; j < len(cfg.logs); j++ {
					if old, oldok := cfg.logs[j][m.Index]; oldok && old != v {
						// some server has already committed a different value for this entry!
						err_msg = fmt.Sprintf("commit index=%v server=%v %v != server=%v %v",
							m.Index, i, m.Command, j, old)
					}
				}
				_, prevok := cfg.logs[i][m.Index-1]
				cfg.logs[i][m.Index] = v
				cfg.mu.Unlock()

				if m.Index > 1 && !prevok {
					err_msg = fmt.Sprintf("server %v apply out of order %v", i, m.Index)
				}
			} else {
				err_msg = fmt.Sprintf("committed command %v is not an int", m.Command)
			}

			if err_msg != "" {
				log.Fatalf("apply error: %v\n", err_msg)
				cfg.applyErr[i] = err_msg
				// keep reading after error so that Raft doesn't block
				// holding locks...
			}
		}
	}()

	rf := Make(ends, i, applyCh)

	cfg.mu.Lock()
	cfg.rafts[i] = rf
	cfg.mu.Unlock()

	svc := rpc.MakeService(rf)
	srv := rpc.MakeServer()
	srv.AddService(svc)
	cfg.net.AddServer(i, srv)
}

func (cfg *config) cleanup() {
	for i := 0; i < len(cfg.rafts); i++ {
		if cfg.rafts[i] != nil {
			cfg.rafts[i].Kill()
		}
	}
	atomic.StoreInt32(&cfg.done, 1)
}

// attach server i to the net.
func (cfg *config) connect(i int) {

	cfg.connected[i] = true

	// outgoing ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.connected[j] {
			endname := cfg.endnames[i][j]
			cfg.net.Enable(endname, true)
		}
	}

	// incoming ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.connected[j] {
			endname := cfg.endnames[j][i]
			cfg.net.Enable(endname, true)
		}
	}
}

// detach server i from the net.
func (cfg *config) disconnect(i int) {

	cfg.connected[i] = false

	// outgoing ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.endnames[i] != nil {
			endname := cfg.endnames[i][j]
			cfg.net.Enable(endname, false)
		}
	}

	// incoming ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.endnames[j] != nil {
			endname := cfg.endnames[j][i]
			cfg.net.Enable(endname, false)
		}
	}
}

func (cfg *config) rpcCount(server int) int {
	return cfg.net.GetCount(server)
}

func (cfg *config) setunreliable(unrel bool) {
	cfg.net.Reliable(!unrel)
}

func (cfg *config) setlongreordering(longrel bool) {
	cfg.net.LongReordering(longrel)
}

// check that there's exactly one leader.
// try a few times in case re-elections are needed.
func (cfg *config) checkOneLeader() int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		leaders := make(map[int][]int)
		for i := 0; i < cfg.n; i++ {
			if cfg.connected[i] {
				if _, t, leader := cfg.rafts[i].GetState(); leader {
					leaders[t] = append(leaders[t], i)
				}
			}
		}

		lastTermWithLeader := -1
		for t, leaders := range leaders {
			if len(leaders) > 1 {
				cfg.t.Fatalf("term %d has %d (>1) leaders", t, len(leaders))
			}
			if t > lastTermWithLeader {
				lastTermWithLeader = t
			}
		}

		if len(leaders) != 0 {
			return leaders[lastTermWithLeader][0]
		}
	}
	cfg.t.Fatalf("expected one leader, got none")
	return -1
}

// check that everyone agrees on the term.
func (cfg *config) checkTerms() int {
	term := -1
	for i := 0; i < cfg.n; i++ {
		if cfg.connected[i] {
			_, xterm, _ := cfg.rafts[i].GetState()
			if term == -1 {
				term = xterm
			} else if term != xterm {
				cfg.t.Fatalf("servers disagree on term")
			}
		}
	}
	return term
}

// check that there's no leader
func (cfg *config) checkNoLeader() {
	for i := 0; i < cfg.n; i++ {
		if cfg.connected[i] {
			_, _, is_leader := cfg.rafts[i].GetState()
			if is_leader {
				cfg.t.Fatalf("expected no leader, but %v claims to be leader", i)
			}
		}
	}
}

// how many servers think a log entry is committed?
func (cfg *config) nCommitted(index int) (int, interface{}) {
	count := 0
	cmd := -1
	for i := 0; i < len(cfg.rafts); i++ {
		if cfg.applyErr[i] != "" {
			cfg.t.Fatal(cfg.applyErr[i])
		}

		cfg.mu.Lock()
		cmd1, ok := cfg.logs[i][index]
		cfg.mu.Unlock()

		if ok {
			if count > 0 && cmd != cmd1 {
				cfg.t.Fatalf("committed values do not match: index %v, %v, %v\n",
					index, cmd, cmd1)
			}
			count += 1
			cmd = cmd1
		}
	}
	return count, cmd
}

// wait for at least n servers to commit.
// but don't wait forever.
func (cfg *config) wait(index int, n int, startTerm int) interface{} {
	to := 10 * time.Millisecond
	for iters := 0; iters < 30; iters++ {
		nd, _ := cfg.nCommitted(index)
		if nd >= n {
			break
		}
		time.Sleep(to)
		if to < time.Second {
			to *= 2
		}
		if startTerm > -1 {
			for _, r := range cfg.rafts {
				if _, t, _ := r.GetState(); t > startTerm {
					// someone has moved on
					// can no longer guarantee that we'll "win"
					return -1
				}
			}
		}
	}
	nd, cmd := cfg.nCommitted(index)
	if nd < n {
		cfg.t.Fatalf("only %d decided for index %d; wanted %d\n",
			nd, index, n)
	}
	return cmd
}

// do a complete agreement.
// it might choose the wrong leader initially,
// and have to re-submit after giving up.
// entirely gives up after about 10 seconds.
// indirectly checks that the servers agree on the
// same value, since nCommitted() checks this,
// as do the threads that read from applyCh.
// returns index.
func (cfg *config) one(cmd int, expectedServers int) int {
	t0 := time.Now()
	starts := 0
	for time.Since(t0).Seconds() < 10 {
		// try all the servers, maybe one is the leader.
		index := -1
		for si := 0; si < cfg.n; si++ {

			starts = (starts + 1) % cfg.n
			var rf *Raft
			cfg.mu.Lock()
			if cfg.connected[starts] {
				rf = cfg.rafts[starts]
			}
			cfg.mu.Unlock()
			if rf != nil {
				index1, _, ok := rf.Start(cmd)
				if ok {
					index = index1
					break
				}
			}
		}

		if index != -1 {
			// somebody claimed to be the leader and to have
			// submitted our command; wait a while for agreement.
			t1 := time.Now()
			for time.Since(t1).Seconds() < 2 {
				nd, cmd1 := cfg.nCommitted(index)
				if nd > 0 && nd >= expectedServers {
					// committed
					if cmd2, ok := cmd1.(int); ok && cmd2 == cmd {
						// and it was the command we submitted.
						return index
					}
				}
				time.Sleep(20 * time.Millisecond)
			}
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
	cfg.t.Fatalf("one(%v) failed to reach agreement", cmd)
	return -1
}
