/**
 * @file SearchForm.tsx
 * @description AND/NOT 多条件検索フォームのReactコンポーネント。
 * Pythonアプリの index.html と results.html のJavaScriptロジックをReactで実装したものです。
 *
 * 機能:
 * - AND/NOT セレクトと入力フィールドの行を動的に追加・削除
 * - スペースキーで自動的に新行を追加
 * - バックスペースキーで空行を削除し、前行にフォーカスを移動
 * - 検索結果ページでは既存クエリをタグとして表示
 */

import { useState, useRef } from 'react';
import type { KeyboardEvent, CompositionEvent, ChangeEvent } from 'react';
import { sendActionLog } from '../api/log';

// ===========================================================================
// 型定義
// ===========================================================================

/**
 * ConditionType は検索条件のタイプ（AND/NOT）を表す型です。
 */
type ConditionType = 'AND' | 'NOT';

/**
 * SearchCondition は1つの検索条件（行）を表す型です。
 */
interface SearchCondition {
    /** React listレンダリング用のユニークキー */
    id: number;
    /** AND: 含む | NOT: 含まない */
    type: ConditionType;
    /** 入力されたキーワード */
    value: string;
}

/**
 * SearchFormProps は SearchForm コンポーネントのpropsを定義します。
 */
interface SearchFormProps {
    /** フォーム送信時に呼ばれるコールバック。最終クエリ文字列を引数に受け取る */
    onSearch: (query: string) => void;
    /** 既存クエリ（検索結果ページでタグとして表示する初期値） */
    initialQuery?: string;
    /** 送信ボタンのラベル */
    submitLabel?: string;
}

// ===========================================================================
// コンポーネント
// ===========================================================================

let conditionKeyCounter = 0; // SearchCondition ID生成用カウンター

/**
 * SearchForm はAND/NOT多条件検索フォームコンポーネントです。
 */
export function SearchForm({ onSearch, initialQuery = '', submitLabel = '検索' }: SearchFormProps) {
    // 既存クエリをタグとして表示する（検索結果ページ用）
    const initialTags = initialQuery
        ? initialQuery
            .replace(/　/g, ' ')
            .split(' ')
            .filter((t) => t.trim() !== '')
        : [];

    const [tags, setTags] = useState<string[]>(initialTags);
    const [conditions, setConditions] = useState<SearchCondition[]>([
        { id: ++conditionKeyCounter, type: 'AND', value: '' },
    ]);

    // 各入力要素への参照（バックスペース時のフォーカス制御用）
    const inputRefs = useRef<Map<number, HTMLInputElement>>(new Map());

    // ===========================================================================
    // タグ操作
    // ===========================================================================

    /** removeTag は指定したタグを削除します */
    const removeTag = (text: string) => {
        setTags((prev) => prev.filter((t) => t !== text));
    };

    // ===========================================================================
    // 条件行操作
    // ===========================================================================

    /** addCondition は新しい空の条件行を追加します */
    const addCondition = () => {
        const newId = ++conditionKeyCounter;
        setConditions((prev) => [...prev, { id: newId, type: 'AND', value: '' }]);
        sendActionLog('add_condition');
        // 追加後にフォーカス（setTimeout でDOM更新を待つ）
        setTimeout(() => {
            inputRefs.current.get(newId)?.focus();
        }, 10);
    };

    /** removeCondition は指定IDの条件行を削除します */
    const removeCondition = (id: number) => {
        setConditions((prev) => {
            const next = prev.filter((c) => c.id !== id);
            // 全行削除された場合は1行追加
            if (next.length === 0 && tags.length === 0) {
                return [{ id: ++conditionKeyCounter, type: 'AND', value: '' }];
            }
            return next;
        });
        sendActionLog('remove_condition');
    };

    /** updateConditionType は指定行のAND/NOTを更新します */
    const updateConditionType = (id: number, type: ConditionType) => {
        setConditions((prev) => prev.map((c) => (c.id === id ? { ...c, type } : c)));
    };

    /** updateConditionValue は指定行の入力値を更新します */
    const updateConditionValue = (id: number, value: string) => {
        setConditions((prev) => prev.map((c) => (c.id === id ? { ...c, value } : c)));
    };

    // ===========================================================================
    // キーボード処理
    // ===========================================================================

    /**
     * handleInputChange はスペースキー入力で自動的に新行を追加します。
     * IMEコンポジション中は無視します（handleCompositionEnd で処理）。
     */
    const handleInputChange = (
        id: number,
        e: ChangeEvent<HTMLInputElement>,
        isComposing: boolean,
    ) => {
        if (isComposing) return;
        const val = e.target.value;
        if (val.endsWith(' ') || val.endsWith('　')) {
            const trimmed = val.trim();
            if (trimmed.length > 0) {
                updateConditionValue(id, trimmed);
                addCondition();
                sendActionLog('auto_add_condition_space');
            } else {
                updateConditionValue(id, '');
            }
        } else {
            updateConditionValue(id, val);
        }
    };

    /**
     * handleCompositionEnd はIME確定後のスペースチェックを行います。
     * 全角スペースで確定した場合の新行追加処理です。
     */
    const handleCompositionEnd = (id: number, e: CompositionEvent<HTMLInputElement>) => {
        const val = (e.target as HTMLInputElement).value;
        if (val.endsWith(' ') || val.endsWith('　')) {
            const trimmed = val.trim();
            if (trimmed.length > 0) {
                updateConditionValue(id, trimmed);
                addCondition();
            } else {
                updateConditionValue(id, '');
            }
        }
    };

    /**
     * handleKeyDown はバックスペースキーで空行を削除し、前行にフォーカスを移動します。
     */
    const handleKeyDown = (id: number, e: KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Backspace') {
            const currentCondition = conditions.find((c) => c.id === id);
            if (currentCondition?.value === '' && conditions.length > 1) {
                const currentIndex = conditions.findIndex((c) => c.id === id);
                if (currentIndex > 0) {
                    e.preventDefault();
                    const prevCondition = conditions[currentIndex - 1];
                    removeCondition(id);
                    setTimeout(() => {
                        inputRefs.current.get(prevCondition.id)?.focus();
                    }, 10);
                    sendActionLog('auto_remove_condition_backspace');
                }
            }
        }
    };

    // ===========================================================================
    // フォーム送信
    // ===========================================================================

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        const queryParts: string[] = [...tags]; // タグの条件を先に追加

        // 各行の条件をクエリに追加
        for (const condition of conditions) {
            const val = condition.value.trim();
            if (val) {
                if (condition.type === 'NOT') {
                    queryParts.push(`-${val}`);
                } else {
                    queryParts.push(val);
                }
            }
        }

        if (queryParts.length === 0) {
            alert('検索キーワードを入力してください。');
            return;
        }

        const finalQuery = queryParts.join(' ');
        sendActionLog('search_submit', { query: finalQuery });
        onSearch(finalQuery);
    };

    // ===========================================================================
    // レンダリング
    // ===========================================================================

    let composingMap: Map<number, boolean> = new Map();

    return (
        <form onSubmit={handleSubmit} className="search-form">
            {/* 既存クエリのタグ表示（検索結果ページでのみ表示） */}
            {tags.length > 0 && (
                <div className="tags-container">
                    {tags.map((tag) => (
                        <div
                            key={tag}
                            className={`tag ${tag.startsWith('-') ? 'tag-not' : ''}`}
                        >
                            <span>{tag.startsWith('-') ? `NOT: ${tag.slice(1)}` : tag}</span>
                            <span className="tag-remove" onClick={() => removeTag(tag)}>
                                ×
                            </span>
                        </div>
                    ))}
                </div>
            )}

            {/* 条件行グループ */}
            <div id="conditionsContainer">
                {conditions.map((condition) => (
                    <div key={condition.id} className="condition-row">
                        {/* AND/NOT セレクト */}
                        <select
                            className="condition-select"
                            value={condition.type}
                            onChange={(e) => updateConditionType(condition.id, e.target.value as ConditionType)}
                        >
                            <option value="AND">AND (含む)</option>
                            <option value="NOT">NOT (含まない)</option>
                        </select>

                        {/* キーワード入力 */}
                        <input
                            type="text"
                            className="search-input"
                            placeholder="キーワード"
                            value={condition.value}
                            ref={(el) => {
                                if (el) inputRefs.current.set(condition.id, el);
                                else inputRefs.current.delete(condition.id);
                            }}
                            onChange={(e) => {
                                const isComposing = composingMap.get(condition.id) ?? false;
                                handleInputChange(condition.id, e, isComposing);
                            }}
                            onCompositionStart={() => composingMap.set(condition.id, true)}
                            onCompositionEnd={(e) => {
                                composingMap.set(condition.id, false);
                                handleCompositionEnd(condition.id, e);
                            }}
                            onKeyDown={(e) => handleKeyDown(condition.id, e)}
                        />

                        {/* 行削除ボタン */}
                        <button
                            type="button"
                            className="remove-btn"
                            onClick={() => removeCondition(condition.id)}
                        >
                            ×
                        </button>
                    </div>
                ))}
            </div>

            <button type="button" className="add-btn" onClick={addCondition}>
                ＋ 条件を追加
            </button>
            <br />
            <button type="submit" className="submit-button">
                {submitLabel}
            </button>
        </form>
    );
}
