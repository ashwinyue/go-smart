package conversation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Manager 对话管理器
type Manager struct {
	stateManager *StateManager
	flow         *ConversationFlow
}

// NewManager 创建新的对话管理器
func NewManager() *Manager {
	return &Manager{
		stateManager: NewStateManager(),
		flow:         NewConversationFlow(),
	}
}

// GetOrCreateState 获取或创建对话状态
func (m *Manager) GetOrCreateState(sessionID, userID string) *ConversationState {
	state, exists := m.stateManager.GetState(sessionID)
	if !exists {
		state = m.stateManager.CreateState(sessionID, userID)
	}
	return state
}

// ProcessMessage 处理用户消息
func (m *Manager) ProcessMessage(ctx context.Context, sessionID, userID, message string) (string, error) {
	// 获取或创建对话状态
	state := m.GetOrCreateState(sessionID, userID)
	
	// 添加用户消息到历史
	err := m.stateManager.AddMessage(sessionID, "user", message, nil)
	if err != nil {
		return "", fmt.Errorf("添加用户消息失败: %w", err)
	}
	
	// 处理输入并获取响应
	response, err := m.flow.ProcessInput(ctx, state, message)
	if err != nil {
		return "", fmt.Errorf("处理消息失败: %w", err)
	}
	
	// 添加助手回复到历史
	err = m.stateManager.AddMessage(sessionID, "assistant", response, nil)
	if err != nil {
		return "", fmt.Errorf("添加助手消息失败: %w", err)
	}
	
	return response, nil
}

// GetConversationHistory 获取对话历史
func (m *Manager) GetConversationHistory(sessionID string) ([]Message, error) {
	state, exists := m.stateManager.GetState(sessionID)
	if !exists {
		return nil, ErrStateNotFound
	}
	
	return state.History, nil
}

// GetCurrentStep 获取当前步骤
func (m *Manager) GetCurrentStep(sessionID string) (string, error) {
	state, exists := m.stateManager.GetState(sessionID)
	if !exists {
		return "", ErrStateNotFound
	}
	
	return state.CurrentStep, nil
}

// GetContext 获取上下文
func (m *Manager) GetContext(sessionID, key string) (interface{}, bool) {
	return m.stateManager.GetContext(sessionID, key)
}

// SetContext 设置上下文
func (m *Manager) SetContext(sessionID string, key string, value interface{}) error {
	return m.stateManager.SetContext(sessionID, key, value)
}

// ResetConversation 重置对话
func (m *Manager) ResetConversation(sessionID string) error {
	state, exists := m.stateManager.GetState(sessionID)
	if !exists {
		return ErrStateNotFound
	}
	
	// 重置状态
	state.CurrentStep = "greeting"
	state.Context = make(map[string]interface{})
	state.History = make([]Message, 0)
	state.LastActivity = time.Now()
	
	return nil
}

// ClearExpiredConversations 清理过期对话
func (m *Manager) ClearExpiredConversations(expiration time.Duration) {
	m.stateManager.ClearExpiredStates(expiration)
}

// RemoveConversation 移除对话
func (m *Manager) RemoveConversation(sessionID string) {
	m.stateManager.RemoveState(sessionID)
}

// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// RegisterStep 注册自定义步骤
func (m *Manager) RegisterStep(step ConversationStep) {
	m.flow.RegisterStep(step)
}