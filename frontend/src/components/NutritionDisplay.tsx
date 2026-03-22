/**
 * @file NutritionDisplay.tsx
 * @description 栄養素情報を表示するReactコンポーネント。
 * レシピ全体の栄養素合計と1人分の栄養素・充足率を表示します。
 */

import type { NutritionInfo, NutritionRatio, NutritionStandards } from '../types/recipe';

/** 数値を小数点1桁でフォーマットするヘルパー関数 */
const fmt1 = (n: number) => n.toFixed(1);
/** 数値を整数でフォーマットするヘルパー関数 */
const fmt0 = (n: number) => Math.round(n).toString();

interface NutritionTotalsProps {
    totals: NutritionInfo;
}

/**
 * NutritionTotals はレシピ全体の栄養素合計を表示するコンポーネントです。
 */
export function NutritionTotals({ totals }: NutritionTotalsProps) {
    return (
        <div className="nutrition-totals">
            <h3>レシピ全体の栄養素 (推定値)</h3>
            <ul>
                <li><span className="label">エネルギー:</span> <span className="value">{fmt1(totals.energy_kcal)} kcal</span></li>
                <li><span className="label">たんぱく質:</span> <span className="value">{fmt1(totals.protein_g)} g</span></li>
                <li><span className="label">脂質:</span> <span className="value">{fmt1(totals.fat_g)} g</span></li>
                <li><span className="label">炭水化物:</span> <span className="value">{fmt1(totals.carbs_g)} g</span></li>
            </ul>
            <ul style={{ marginTop: '10px', borderTop: '1px dashed #ccc', paddingTop: '10px' }}>
                <li><span className="label">食物繊維:</span> <span className="value">{fmt1(totals.fiber_g)} g</span></li>
                <li><span className="label">食塩相当量:</span> <span className="value">{fmt1(totals.salt_g)} g</span></li>
            </ul>
        </div>
    );
}

interface NutritionPerServingProps {
    perServing: NutritionInfo;
    ratios: NutritionRatio;
    standards: NutritionStandards;
    servingSize: number;
}

/**
 * NutritionPerServing は1人分の栄養素と基準充足率を表示するコンポーネントです。
 */
export function NutritionPerServing({ perServing, ratios, standards, servingSize }: NutritionPerServingProps) {
    return (
        <div className="nutrition-totals">
            <h3>1人分の栄養素 ({servingSize}人分として計算)</h3>
            <ul>
                <li>
                    <span className="label">エネルギー:</span>{' '}
                    <span className="value">{fmt1(perServing.energy_kcal)} kcal</span>{' '}
                    <span className="small-note">({fmt0(ratios.energy)}%) 基準: {standards.energy_kcal}kcal</span>
                </li>
                <li>
                    <span className="label">たんぱく質:</span>{' '}
                    <span className="value">{fmt1(perServing.protein_g)} g</span>{' '}
                    <span className="small-note">({fmt0(ratios.protein)}%) 基準: {standards.protein_g}g</span>
                </li>
                <li>
                    <span className="label">脂質:</span>{' '}
                    <span className="value">{fmt1(perServing.fat_g)} g</span>{' '}
                    <span className="small-note">({fmt0(ratios.fat)}%) 基準: {standards.fat_g}g</span>
                </li>
                <li>
                    <span className="label">炭水化物:</span>{' '}
                    <span className="value">{fmt1(perServing.carbs_g)} g</span>{' '}
                    <span className="small-note">({fmt0(ratios.carbs)}%) 基準: {standards.carbs_g}g</span>
                </li>
            </ul>
            <ul style={{ marginTop: '10px', borderTop: '1px dashed #ccc', paddingTop: '10px' }}>
                <li>
                    <span className="label">食物繊維:</span>{' '}
                    <span className="value">{fmt1(perServing.fiber_g)} g</span>{' '}
                    <span className="small-note">({fmt0(ratios.fiber)}%) 基準: {standards.fiber_g}g</span>
                </li>
                <li>
                    <span className="label">食塩相当量:</span>{' '}
                    <span className="value">{fmt1(perServing.salt_g)} g</span>{' '}
                    <span className="small-note">({fmt0(ratios.salt)}%) 基準: {standards.salt_g}g</span>
                </li>
            </ul>
            <p style={{ fontSize: '0.8em', color: '#888', textAlign: 'right', marginTop: '5px' }}>
                ※カッコ内は基準摂取量に対する充足率
            </p>
        </div>
    );
}
