package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 3D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 3D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type LogEntry struct{
	Command interface{}
	Index 	int
	Term 	int
}

type NodeState int

const (
    Follower NodeState = iota // 从 0 开始
    Candidate             // 自动递增为 1
    Leader                // 自动递增为 2
)

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.RWMutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (3A, 3B, 3C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	//3A
	currentTerm int 
	votefor 	int
	state       NodeState // 0 -follower   1 - conditate  2 - leader
	timestamp   time.Time
	applycond   *sync.Cond
	applychan	chan ApplyMsg
	//不需要持久保存的易失状态，丢失同步Leader就行
	logs 		[]LogEntry
	commitIndex int //已提交的最高日志项的索引
	lastApplied int //应用于状态机的最高日志项的索引
	//对领导者来说有意义的易失状态，变更leader后初始化
	nextIndex	[]int // 对每一个追随者，下一个要发送给该服务器的日志条目的索引
	matchIndex	[]int // 对每一个追随者,已知复制到该服务器的最高日志条目的索引

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (3A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term  = rf.currentTerm
	isleader = (rf.state == Leader)
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}


// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}


// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}


// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	 Term			int // 候选人的任期
	 CandidateId	int 
	 LastLogIndex 	int //  候选人最后一次日志条目的索引,选候选人的时候要比较谁的log全
	 LastLogTerm	int //  候选人最后一次日志条目的任期
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	Term  		int  //用于候选者更新自身
	VoteGranted bool
}

type AppendEntries struct{
	Term 		int 
	LeaderId 	int
	//为了保障节点之间的日志一致性（如果有节点故障了几个Term，
	//通过这个可以检查出来，并且完成和Leader同步）
	PreLogIndex int // 紧接在新条目之前的日志的索引
	PreLogTerm  int 
	Entries []LogEntry //空的就是heartbeat
	LeaderCommit	  int  // leader's commitindex
}

type AppendEntriesreply struct{
	Term   int 
	Success bool
}

// example RequestVote RPC handler.
// handler : 暗示了这个函数用来响应其他函数的请求
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3A, 3B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	Debug(dVote,"[%v]receried the request vote from [%v], Term : [%v]",rf.me,args.CandidateId,rf.currentTerm)
	
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}else if args.Term >  rf.currentTerm{
		rf.currentTerm = args.Term
		rf.state = Follower
		rf.votefor = -1
	}

	if rf.votefor == -1 || rf.votefor == args.CandidateId{
		lastLogTerm := rf.logs[len(rf.logs)-1].Term
		if args.LastLogTerm > lastLogTerm || 
		   (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= len(rf.logs)-1) {
			reply.Term = args.Term
			reply.VoteGranted = true
			rf.votefor = args.CandidateId
			rf.timestamp = time.Now()
			Debug(dVote,"[term : %v] : [%v] vote for  [%v]",args.Term,rf.me,args.CandidateId)
			return
		}
	}
	reply.VoteGranted = false
}

// received heartbeat handler
func (rf *Raft)AppendEntry(args *AppendEntries, reply *AppendEntriesreply){
	rf.mu.Lock()
	defer rf.mu.Unlock()
	Debug(dLog2,"[term %d]: Raft[%d] [state %d] receive AppendEntries from Raft[%d]", rf.currentTerm, rf.me, rf.state, args.LeaderId)
	
	if args.Term < rf.currentTerm{
		reply.Term = rf.currentTerm
		Debug(dTerm,"receiver [%v] has bigger term", rf.me)
		reply.Success = false
		return
	}else if(args.Term > rf.currentTerm){
		rf.state = Follower
		rf.currentTerm = args.Term
		rf.votefor = -1 // haven't vote yet
	}
	
	reply.Success = true
	if(args.PreLogIndex >= len(rf.logs) ||rf.logs[args.PreLogIndex].Term != args.PreLogTerm){
		reply.Success = false
		Debug(dLog2,"[%v] The log is inconsistent with the leader",rf.me)
		return
	}

	for i :=0; i < len(args.Entries); i++{
		index := args.Entries[i].Index
		if(index <= len(rf.logs)-1 && rf.logs[index].Term != args.Entries[i].Term){
			Debug(dLog2,"[%v] logs index [%v] conflicts with new one",rf.me,index)
			rf.logs = rf.logs[:index]	
		}
		rf.logs = append(rf.logs, args.Entries[i])
	}

	if(args.LeaderCommit > rf.commitIndex){
		rf.commitIndex = min(args.LeaderCommit,len(rf.logs)-1)
		if rf.commitIndex > rf.lastApplied{
			rf.applycond.Broadcast()
		}
	}

	reply.Term = args.Term
	rf.timestamp = time.Now()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (rf *Raft)checkAppendresult(server int, args *AppendEntries) bool{
	Debug(dLog,"[term %d]:Raft [%d] [state %d] sends appendentries RPC to server[%d]", rf.currentTerm, rf.me, rf.state, server)
	reply := AppendEntriesreply{}
	ok := rf.sendAppendEntries(server,args,&reply)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if !ok{
		Debug(dLog2,"[%v] term :[%v] received AppendEntries form [%v] Term :[%v] unsuccess",server,reply.Term,rf.me,rf.currentTerm)
	}

	if reply.Term > rf.currentTerm {
		rf.currentTerm = reply.Term
		rf.state = Follower
		rf.timestamp = time.Now()
		rf.votefor = -1
	}
	return reply.Success
}

//检查有客户端有没有消息发送到Leader，如果有且大于等于服务器的nextIndex则发送带有日志的AppendEntries
func (rf *Raft) checknewEntry(server int){
	newargs := AppendEntries{}
	for{	
		rf.mu.Lock()
		if rf.state != Leader || rf.killed(){
			rf.mu.Unlock()
			return
		}
		lastIndex := len(rf.logs)-1
		nextindex := rf.nextIndex[server]
		newargs.Entries = rf.logs[nextindex:]
		newargs.PreLogIndex = nextindex-1
		newargs.PreLogTerm = rf.logs[nextindex-1].Term
		newargs.LeaderCommit = rf.commitIndex
		newargs.Term = rf.currentTerm
		rf.mu.Unlock()

		if(lastIndex >= nextindex){
			Debug(dLog,"[Raft %v], send a new log to server[%v]",rf.me,server)
			result := rf.checkAppendresult(server,&newargs)
			rf.mu.Lock()
			// 出现了term confusion
			if newargs.Term != rf.currentTerm{
				rf.mu.Unlock()
				return
			}
			if result {
				rf.nextIndex[server] = lastIndex+1
				rf.matchIndex[server] = lastIndex
				Debug(dLog,"[term %d]: Raft[%d] successfully append entries to Raft[%d]", rf.currentTerm, rf.me, server)
			}else{
				rf.nextIndex[server] = int(math.Max(1.0, float64(rf.nextIndex[server]-1)))
				rf.mu.Unlock()
				continue
			}
			rf.mu.Unlock()
		}
		time.Sleep(time.Millisecond * time.Duration(10))
	}
}

func (rf *Raft) checkcommit(){
	for{
		rf.mu.Lock()
		if rf.killed() || rf.state != Leader{
			rf.mu.Unlock()
			return
		}

		if len(rf.logs)-1 > rf.commitIndex{
				termcommit := rf.commitIndex+1
				// 找到最新的可提交的索引,每次只提交一个
				sum := 0
				for i := 0; i < len(rf.peers); i++ {
					if i == rf.me{
						continue
					}
					
					if rf.matchIndex[i] >= termcommit {
							sum++
					}
				}

				if sum > len(rf.peers)/2 && rf.logs[termcommit].Term == rf.currentTerm{
					Debug(dCommit,"term : [%v]  Raft :[%v] has change the commitx successfully",rf.currentTerm,rf.me)
					rf.commitIndex = termcommit
					rf.applycond.Broadcast()
				}
			}
		rf.mu.Unlock()
		time.Sleep(time.Millisecond * time.Duration(10))
	}
}

func (rf *Raft) startsendAppend(){
	timeout := time.Duration(150) *time.Millisecond

	for rf.killed() == false{
		if(rf.state == Leader){
			rf.mu.Lock()
			args := AppendEntries{
					Term: rf.currentTerm,
					LeaderId: rf.me,
					PreLogIndex: len(rf.logs)-1,
					PreLogTerm: rf.logs[len(rf.logs)-1].Term,
					Entries: make([]LogEntry,0),
					LeaderCommit: rf.commitIndex,
				}	
			rf.mu.Unlock()	
			

			for server := range rf.peers{
				if server == rf.me{
					continue
				}
				go rf.checkAppendresult(server,&args)
			}
		}
		time.Sleep(timeout)
		
	}
}
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntries, reply *AppendEntriesreply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntry", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (3B).
	rf.mu.Lock()

	isLeader = (rf.state == Leader)
	if !isLeader || rf.killed() {
		rf.mu.Unlock()
		return index,term,false
	}
	Debug(dClient,"[term %d]: Raft [%d] start recerived new log", rf.currentTerm, rf.me)
	index = len(rf.logs)
	term = rf.currentTerm
	rf.logs = append(rf.logs,LogEntry{Command: command,
									  Index : index,
									  Term : term })
	rf.mu.Unlock()

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for !rf.killed(){
		// Your code here (3A)
		// Check if a leader election should be started.
		//注意Condidate也可能要重新发起选举
		timeout := ElectionRandTime()
		time.Sleep(timeout)
		
		rf.mu.Lock()
		if rf.state != Leader && time.Since(rf.timestamp) >= timeout{
			go rf.startElection()	
		}		
		// pause for a random amount of time between 50 and 350
		// milliseconds.
		rf.mu.Unlock()
		ms := 50 + (rand.Int63() % 300)
		time.Sleep(time.Duration(ms) * time.Millisecond)		
	}
}

func (rf *Raft)startElection(){
	rf.mu.Lock()
	rf.state = Candidate
	rf.currentTerm += 1
	rf.timestamp = time.Now()
	rf.votefor = rf.me

	Debug(dVote,"[%v] start election,Term:[%v]",rf.me,rf.currentTerm)
	
	// send voterequest to others
	votenum := 1
	args := RequestVoteArgs{
		Term : rf.currentTerm,
		CandidateId: rf.me,
		LastLogIndex : len(rf.logs) -1,
		LastLogTerm:  rf.logs[len(rf.logs) -1].Term,
	}
	rf.mu.Unlock()

	var votemux sync.Mutex
	// 避免重复的接受投票操作
	electionfinished := false

	for server := range rf.peers{
		if server == rf.me{
			Debug(dVote,"[%v] vote for himself",rf.me)
			continue
		}
		go func(server int){

			result := rf.CheckVoteResult(server,&args)

			if !result && rf.currentTerm > args.Term{
				return
			} 
			votemux.Lock()
			if result && !electionfinished {
				votenum++
				Debug(dVote,"[%v] vote for the condidate [%v]",server,rf.me)	
				if votenum > len(rf.peers)/2{
					electionfinished = true
					rf.mu.Lock()
					rf.state = Leader
					for i:= 0; i < len(rf.peers); i++{
						rf.nextIndex[i] = len(rf.logs)
						rf.matchIndex[i] = 0
					}
					rf.mu.Unlock()
					Debug(dLeader,"[%v] become new Leader",rf.me)
					go rf.startsendAppend()
					go rf.allocateAppendCheckers()
					go rf.checkcommit()
				}
			}
			votemux.Unlock()
		}(server)
	}
}

func (rf *Raft) allocateAppendCheckers(){
	for i:=0; i < len(rf.peers); i++{
		if i == rf.me{
			continue
		}
		go rf.checknewEntry(i)
	}
}
func (rf *Raft) CheckVoteResult(server int,args *RequestVoteArgs) bool{
	reply := RequestVoteReply{}
	ok := rf.sendRequestVote(server,args,&reply)
	if !ok{
		return false
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// 当两个Condidate同时请求投票的时候，万一别人快一点已经完成了投票，此时我们放弃投票。
	if args.Term < rf.currentTerm{
		return false
	}

	if reply.Term > args.Term{
		Debug(dVote," condidate [%v] Term :[%v] recerived higher term [%v] from others ",args.CandidateId,args.Term,reply.Term)
		rf.state = Follower
		rf.currentTerm = reply.Term
		rf.votefor = -1
		return false
	}
	return reply.VoteGranted
}

func (rf *Raft)checkApply(){
	for !rf.killed(){
		rf.mu.Lock()
	// 信号量的用法就是需要持有对应的锁，不会造成死锁。
		for rf.lastApplied >= rf.commitIndex{
			rf.applycond.Wait()
		}
		
		rf.lastApplied++
		Debug(dCommit,"[Term : %v],Raft :%v [state %d] ready to apply the new log [%v] to statemachine,",rf.currentTerm,rf.me,rf.state,rf.lastApplied)
		Msg := ApplyMsg{
			CommandValid: true,
			Command: rf.logs[rf.lastApplied].Command,
			CommandIndex: rf.lastApplied,
		}
		rf.mu.Unlock()
		rf.applychan <- Msg
		Debug(dCommit,"[Term : %v] [Raft :%v]  [state %d] send new log to statemachine successfully",rf.currentTerm,rf.me,rf.state)
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func ElectionRandTime() time.Duration{
	ms := 300 + (rand.Int63() % 100)
	return time.Duration(ms) * time.Millisecond
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3A, 3B, 3C).
	length := len(peers)
	rf.logs = make([]LogEntry, 0)
	rf.logs = append(rf.logs, LogEntry{Term: 0})
	rf.nextIndex = make([]int,length)
	rf.matchIndex = make([]int,length)
	rf.timestamp = time.Now()
	rf.state = Follower
	rf.currentTerm = 0
	rf.votefor = -1
	rf.applycond = sync.NewCond(&rf.mu)
	rf.applychan = applyCh
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()
	// 开启一个goroutine来检测是否有需要应用到状态机的日志
	go rf.checkApply()
	return rf
}
