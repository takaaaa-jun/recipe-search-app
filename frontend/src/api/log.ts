/**
 * @file log.ts
 * @description ユーザーアクションログAPIのクライアント関数。
 * フロントエンド操作（検索実行、条件追加など）をバックエンドに送信します。
 */

const API_BASE = '/api';

/**
 * LogActionPayload はログAPIに送信するデータの型です。
 */
interface LogActionPayload {
    /** アクション名（例: "search_submit", "add_condition", "view_recipe"） */
    action: string;
    /** アクションの補足情報（検索クエリなど任意のフィールド） */
    details?: Record<string, unknown>;
    /** アクションが発生したページのURL */
    url?: string;
    /** アクション発生時刻（ISO 8601形式） */
    timestamp?: string;
}

/**
 * sendActionLog はユーザーアクションをバックエンドのログAPIに送信します。
 * fire-and-forget方式のため、エラーはコンソールに出力するだけで、
 * 呼び出し元のUXには影響を与えません。
 *
 * @param action - アクション名
 * @param details - 補足情報（任意）
 */
export async function sendActionLog(
    action: string,
    details?: Record<string, unknown>,
): Promise<void> {
    const payload: LogActionPayload = {
        action,
        details,
        url: window.location.href,
        timestamp: new Date().toISOString(),
    };

    try {
        await fetch(`${API_BASE}/log_action`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
    } catch (error) {
        // ログ送信の失敗はサイレントに処理（UXに影響させない）
        console.error('Action log failed:', error);
    }
}
