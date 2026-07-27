package dao

import (
	"Main/model"
	"context"
	"log"
	"time"
)

func StartCleanChat(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	log.Println("自动清理任务已启动，每小时检查一次过期聊天记录")
	for {
		select {
		case <-ctx.Done():
			log.Println("清理任务已停止")
			return
		case <-ticker.C:
			cleanExpiredChat()
		}
	}
}

// cleanExpiredChat 清理过期记录
func cleanExpiredChat() {
	result := dataBase.Unscoped().Where(
		"updated_at < ?", time.Now().AddDate(0, -3, 0),
	).Delete(&model.ChatModel{})
	if result.Error != nil {
		log.Printf("清理失败: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("成功清理%d条过期对话", result.RowsAffected)
	}
}
