package conversation

import (
	"fmt"
	"sync"
	"time"
)

// ConversationState 对话状态
type ConversationState struct {
	SessionID      string                 `json:"session_id"`
	UserID         string                 `json:"user_id"`
	CurrentStep    string                 `json:"current_step"`    // 当前对话步骤
	Context        map[string]interface{} `json:"context"`        // 对话上下文
	History        []Message              `json:"history"`        // 对话历史
	LastActivity   time.Time              `json:"last_activity"`  // 最后活动时间
	CreatedAt      time.Time              `json:"created_at"`     // 创建时间
}

// Message 对话消息
type Message struct {
	Role      string                 `json:"role"`      // 角色: user, assistant, system
	Content   string                 `json:"content"`   // 消息内容
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// StateManager 对话状态管理器
type StateManager struct {
	states map[string]*ConversationState // 会话ID -> 状态
	mutex  sync.RWMutex                   // 读写锁
}

// NewStateManager 创建新的状态管理器
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]*ConversationState),
	}
}

// GetState 获取对话状态
func (sm *StateManager) GetState(sessionID string) (*ConversationState, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	state, exists := sm.states[sessionID]
	return state, exists
}

// CreateState 创建新的对话状态
func (sm *StateManager) CreateState(sessionID, userID string) *ConversationState {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	now := time.Now()
	state := &ConversationState{
		SessionID:    sessionID,
		UserID:       userID,
		CurrentStep:  "greeting",
		Context:      make(map[string]interface{}),
		History:      make([]Message, 0),
		LastActivity: now,
		CreatedAt:    now,
	}
	
	sm.states[sessionID] = state
	return state
}

// UpdateState 更新对话状态
func (sm *StateManager) UpdateState(sessionID string, updateFunc func(*ConversationState)) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	state, exists := sm.states[sessionID]
	if !exists {
		return ErrStateNotFound
	}
	
	updateFunc(state)
	state.LastActivity = time.Now()
	return nil
}

// AddMessage 添加消息到对话历史
func (sm *StateManager) AddMessage(sessionID string, role, content string, metadata map[string]interface{}) error {
	return sm.UpdateState(sessionID, func(state *ConversationState) {
		message := Message{
			Role:      role,
			Content:   content,
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		state.History = append(state.History, message)
	})
}

// SetCurrentStep 设置当前步骤
func (sm *StateManager) SetCurrentStep(sessionID, step string) error {
	return sm.UpdateState(sessionID, func(state *ConversationState) {
		state.CurrentStep = step
	})
}

// SetContext 设置上下文
func (sm *StateManager) SetContext(sessionID string, key string, value interface{}) error {
	return sm.UpdateState(sessionID, func(state *ConversationState) {
		state.Context[key] = value
	})
}

// GetContext 获取上下文
func (sm *StateManager) GetContext(sessionID, key string) (interface{}, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	state, exists := sm.states[sessionID]
	if !exists {
		return nil, false
	}
	
	value, exists := state.Context[key]
	return value, exists
}

// ClearExpiredStates 清理过期状态
func (sm *StateManager) ClearExpiredStates(expiration time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	now := time.Now()
	for sessionID, state := range sm.states {
		if now.Sub(state.LastActivity) > expiration {
			delete(sm.states, sessionID)
		}
	}
}

// RemoveState 移除对话状态
func (sm *StateManager) RemoveState(sessionID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	delete(sm.states, sessionID)
}

// 错误定义
var (
	ErrStateNotFound = fmt.Errorf("conversation state not found")
)