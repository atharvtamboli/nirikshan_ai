package main
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"github.com/gin-gonic/gin"
	"os"
)
type TxStatus string
const (
	StatusProposed TxStatus = "PROPOSED"
	StatusApproved TxStatus = "APPROVED"
	StatusActive   TxStatus = "ACTIVE"
	StatusAudit    TxStatus = "UNDER_AUDIT"
	StatusHalted   TxStatus = "HALTED"
)
type AuditLog struct {
	Timestamp string `json:"timestamp"`
	Endorser  string `json:"endorser"`
	Action    string `json:"action"`
	Details   string `json:"details"`
}
type ProjectData struct {
	ProjectID   string   `json:"project_id"`
	Title       string   `json:"title"`
	Department  string   `json:"department"`
	Budget      float64  `json:"budget"`
	Utilized    float64  `json:"utilized"`
	Status      TxStatus `json:"status"`
	Contractor  string   `json:"contractor"`
	Coordinates string   `json:"coordinates"`
}
type Block struct {
	Index        int         `json:"index"`
	Timestamp    string      `json:"timestamp"`
	Payload      ProjectData `json:"payload"`
	AuditHistory []AuditLog  `json:"audit_history"`
	PrevHash     string      `json:"prev_hash"`
	Hash         string      `json:"hash"`
	Nonce        int         `json:"nonce"`
}
type Blockchain struct {
	Chain []Block
	mu    sync.RWMutex
}
var GovChain = &Blockchain{}
func calculateHash(b Block) string {
	payloadBytes, _ := json.Marshal(b.Payload)
	auditBytes, _ := json.Marshal(b.AuditHistory)
	record := fmt.Sprintf("%d%s%s%s%s%d", b.Index, b.Timestamp, string(payloadBytes), string(auditBytes), b.PrevHash, b.Nonce)
	h := sha256.Sum256([]byte(record))
	return hex.EncodeToString(h[:])
}
func mineBlock(b *Block, difficulty int) {
	target := ""
	for i := 0; i < difficulty; i++ { target += "0" }
	for {
		b.Hash = calculateHash(*b)
		if b.Hash[:difficulty] == target { break }
		b.Nonce++
	}
}
func (bc *Blockchain) InitGenesis() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	genesis := Block{
		Index:     0,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload:   ProjectData{ProjectID: "SYS_GENESIS", Title: "Network_Initialization", Status: StatusActive},
		PrevHash:  "0000000000000000000000000000000000000000000000000000000000000000",
	}
	mineBlock(&genesis, 2)
	bc.Chain = append(bc.Chain, genesis)
}
func (bc *Blockchain) SaveToFile() {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	data, _ := json.MarshalIndent(bc.Chain, "", "  ")
	os.WriteFile("ledger_backup.json", data, 0644)
}

func (bc *Blockchain) LoadFromFile() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	data, err := os.ReadFile("ledger_backup.json")
	if err != nil {
		return false
	}
	json.Unmarshal(data, &bc.Chain)
	return true
}
func main() {
	if !GovChain.LoadFromFile() {
		GovChain.InitGenesis()
		GovChain.SaveToFile()
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.GET("/", func(c *gin.Context) { c.File("index.html") })
	r.GET("/api/ledger", func(c *gin.Context) {
		GovChain.mu.RLock()
		defer GovChain.mu.RUnlock()
		c.JSON(http.StatusOK, GovChain.Chain)
	})
	r.POST("/api/project", func(c *gin.Context) {
		var req struct {
			Payload  ProjectData `json:"payload"`
			Endorser string      `json:"endorser"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON Schema"})
			return
		}
		GovChain.mu.Lock()
		defer GovChain.mu.Unlock()
		prev := GovChain.Chain[len(GovChain.Chain)-1]
		newBlock := Block{
			Index:     prev.Index + 1,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   req.Payload,
			PrevHash:  prev.Hash,
			AuditHistory: []AuditLog{{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Endorser:  req.Endorser,
				Action:    "PROPOSAL",
				Details:   "Project entry",
			}},
		}
		mineBlock(&newBlock, 2)
		GovChain.Chain = append(GovChain.Chain, newBlock)
		c.JSON(201, newBlock)
	})
	r.POST("/api/audit", func(c *gin.Context) {
		var req struct {
			ProjectID string   `json:"project_id"`
			FundDelta float64  `json:"fund_delta"`
			Status    TxStatus `json:"status"`
			Endorser  string   `json:"endorser"`
			Details   string   `json:"details"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid Audit Schema"})
			return
		}
		GovChain.mu.Lock()
		defer GovChain.mu.Unlock()
		var lastState ProjectData
		var history []AuditLog
		found := false
		for i := len(GovChain.Chain) - 1; i >= 0; i-- {
			if GovChain.Chain[i].Payload.ProjectID == req.ProjectID {
				lastState = GovChain.Chain[i].Payload
				history = append([]AuditLog{}, GovChain.Chain[i].AuditHistory...)
				found = true
				break
			}
		}
		if !found {
			c.JSON(404, gin.H{"error": "Project not found"})
			return
		}
		lastState.Utilized += req.FundDelta
		lastState.Status = req.Status
		prev := GovChain.Chain[len(GovChain.Chain)-1]
		history = append(history, AuditLog{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Endorser:  req.Endorser,
			Action:    "AUDIT",
			Details:   req.Details,
		})
		newBlock := Block{
			Index: prev.Index + 1, Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload: lastState, AuditHistory: history, PrevHash: prev.Hash,
		}
		mineBlock(&newBlock, 2)
		GovChain.Chain = append(GovChain.Chain, newBlock)
		c.JSON(200, newBlock)
	})
	r.Run(":8080")
}
