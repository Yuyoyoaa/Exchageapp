<template>
  <div class="exchange-wrapper" v-loading="loading">
    <el-card class="exchange-card">
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <el-icon :size="24"><Money /></el-icon>
            <h2>实时货币兑换计算器</h2>
          </div>
          <div>
            <el-tag type="info" effect="plain" size="large">数据实时更新</el-tag>
          </div>
        </div>
      </template>

      <div class="card-body-layout">
        <!-- 左侧：计算核心区 -->
        <div class="exchange-core">
          <el-form :model="form" label-position="top" class="exchange-form" size="large">
            <el-row :gutter="20" align="middle" class="form-row">
              <el-col :xs="24" :sm="10">
                <el-form-item label="持有货币">
                  <el-select 
                    v-model="form.fromCurrency" 
                    filterable
                    style="width: 100%;" 
                    @change="handleCurrencyChange"
                  >
                    <el-option v-for="c in currencies" :key="c" :value="c" :label="c" />
                  </el-select>
                </el-form-item>
              </el-col>

              <el-col :xs="24" :sm="4" class="swap-container">
                <el-button 
                  type="primary" 
                  circle 
                  class="swap-btn"
                  @click="swapCurrencies"
                >
                  <el-icon><Switch /></el-icon>
                </el-button>
              </el-col>

              <el-col :xs="24" :sm="10">
                <el-form-item label="目标货币">
                  <el-select 
                    v-model="form.toCurrency" 
                    filterable
                    style="width: 100%;" 
                    @change="handleCurrencyChange"
                  >
                    <el-option v-for="c in currencies" :key="c" :value="c" :label="c" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="兑换金额" style="margin-top: 20px;">
              <el-input 
                v-model.number="form.amount" 
                type="number" 
                placeholder="请输入金额"
                class="amount-input"
                @input="calculateResult"
              >
                <template #prepend>{{ form.fromCurrency }}</template>
              </el-input>
            </el-form-item>

            <div class="result-box">
              <div v-if="result !== null" class="result-text">
                <span class="approx">≈</span> {{ result.toFixed(2) }} <span class="currency-unit">{{ form.toCurrency }}</span>
              </div>
              <div v-else class="result-placeholder">请输入金额查看结果</div>
              <div class="rate-hint">
                当前汇率：1 {{ form.fromCurrency }} ≈ {{ currentRate }} {{ form.toCurrency }}
              </div>
            </div>

            <el-button type="primary" class="exchange-btn" @click="exchange">
              立即兑换
            </el-button>
          </el-form>
        </div>

        <!-- 右侧：关联汇率行情 -->
        <div class="rates-panel">
          <div class="panel-title">
            <el-icon><TrendCharts /></el-icon> 
            {{ form.fromCurrency }} 兑换其他主要货币汇率
          </div>
          <el-table 
            :data="relatedRates" 
            size="default"
            :show-header="true"
            stripe
            height="450"
            style="width: 100%;"
          >
            <el-table-column prop="toCurrency" label="目标货币" />
            <el-table-column prop="rate" label="汇率" align="right">
                <template #default="scope">
                  <span style="color: #E6A23C; font-weight: bold;">{{ scope.row.rate.toFixed(4) }}</span>
                </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {
  Money,
  Switch, TrendCharts
} from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { computed, onMounted, ref } from 'vue';
import axios from '../axios';

interface ExchangeRate {
  id?: number;
  fromCurrency: string;
  toCurrency: string;
  rate: number; // 后端返回的是 USD -> X 的汇率
  date?: string;
}

const form = ref({ fromCurrency: 'USD', toCurrency: 'CNY', amount: 100 });
const baseRates = ref<ExchangeRate[]>([]); // 存储后端返回的基准汇率列表
const currencies = ref<string[]>([]);
const result = ref<number | null>(null);
const loading = ref(false);

const fetchExchangeRates = async () => {
  loading.value = true;
  try {
    // 请求后端的新接口 (GetBaseRates)，获取所有基准汇率
    const res = await axios.get('/exchangeRates');
    baseRates.value = res.data;

    // 提取所有可用货币
    const curSet = new Set<string>();
    curSet.add('USD'); // 确保基准货币 USD 存在
    baseRates.value.forEach(r => { 
      curSet.add(r.toCurrency); // 后端数据全是 USD -> X，取 toCurrency
    });
    
    currencies.value = [...curSet].sort((a, b) => a.localeCompare(b));

    // 初始化默认选中值校验
    if (!currencies.value.includes(form.value.fromCurrency) && currencies.value.length > 0) {
        form.value.fromCurrency = currencies.value[0];
    }
    if (!currencies.value.includes(form.value.toCurrency) && currencies.value.length > 1) {
        // 避免目标和源相同
        form.value.toCurrency = currencies.value.find(c => c !== form.value.fromCurrency) || currencies.value[1];
    }

    calculateResult();
  } catch (err) { 
    console.error(err);
    ElMessage.error("获取汇率数据失败"); 
  } finally {
    loading.value = false;
  }
};

// 核心逻辑：前端计算交叉汇率
// 公式：Rate(A->B) = Rate(USD->B) / Rate(USD->A)
const getCrossRate = (from: string, to: string): number | null => {
  if (from === to) return 1;

  // 1. 获取 USD -> From 的汇率
  let rateUSDToFrom = 1;
  if (from !== 'USD') {
    const r = baseRates.value.find(i => i.toCurrency === from);
    if (!r) return null; // 数据缺失
    rateUSDToFrom = r.rate;
  }

  // 2. 获取 USD -> To 的汇率
  let rateUSDToTo = 1;
  if (to !== 'USD') {
    const r = baseRates.value.find(i => i.toCurrency === to);
    if (!r) return null;
    rateUSDToTo = r.rate;
  }

  return rateUSDToTo / rateUSDToFrom;
};

const calculateResult = () => {
  if (!form.value.amount || typeof form.value.amount !== 'number' || form.value.amount <= 0) { 
      result.value = null; 
      return; 
  }
  
  const rate = getCrossRate(form.value.fromCurrency, form.value.toCurrency);
  result.value = (rate !== null) ? form.value.amount * rate : null;
};

const handleCurrencyChange = () => {
  calculateResult();
};

const currentRate = computed(() => {
  const rate = getCrossRate(form.value.fromCurrency, form.value.toCurrency);
  return rate !== null ? rate.toFixed(4) : '---';
});

const relatedRates = computed(() => {
  if (currencies.value.length === 0) return [];
  
  const list: { toCurrency: string; rate: number }[] = [];
  
  // 常用货币列表，用于排序优化体验
  const popularCurrencies = ['USD', 'CNY', 'EUR', 'GBP', 'JPY', 'HKD', 'AUD', 'CAD'];

  currencies.value.forEach(target => {
    if (target === form.value.fromCurrency) return; // 排除自身
    
    const rate = getCrossRate(form.value.fromCurrency, target);
    if (rate !== null) {
      list.push({ toCurrency: target, rate });
    }
  });

  // 排序：热门货币在前，其他按字母排序
  return list.sort((a, b) => {
    const idxA = popularCurrencies.indexOf(a.toCurrency);
    const idxB = popularCurrencies.indexOf(b.toCurrency);
    if (idxA !== -1 && idxB !== -1) return idxA - idxB;
    if (idxA !== -1) return -1;
    if (idxB !== -1) return 1;
    return a.toCurrency.localeCompare(b.toCurrency);
  });
});

const swapCurrencies = () => {
  [form.value.fromCurrency, form.value.toCurrency] = [form.value.toCurrency, form.value.fromCurrency];
  calculateResult();
};

const exchange = () => {
  if (!form.value.amount) return ElMessage.warning("请输入金额");
  if (result.value !== null) {
    ElMessage.success(`兑换成功！预计获得 ${result.value.toFixed(2)} ${form.value.toCurrency}`);
  } else {
    ElMessage.error("暂不支持该货币组合兑换");
  }
};

onMounted(fetchExchangeRates);
</script>

<style scoped>
.exchange-wrapper {
  min-height: calc(100vh - 60px);
  background: #f5f7fa;
  display: flex;
  justify-content: center;
  align-items: flex-start; /* 改为顶部对齐，适应不同高度 */
  padding: 40px 20px;
  box-sizing: border-box;
}

.exchange-card {
  width: 100%;
  max-width: 1200px;
  min-height: 600px;
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #303133;
}
.header-title h2 { margin: 0; font-size: 22px; font-weight: 600; }

.card-body-layout {
  display: flex;
  gap: 40px;
  padding: 20px 0;
}

.exchange-core {
  flex: 1.4; /* 左侧略宽 */
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  padding-top: 10px;
}

.rates-panel {
  flex: 1;
  background: #fcfcfc;
  border-radius: 12px;
  padding: 24px;
  border: 1px solid #ebeef5;
  display: flex;
  flex-direction: column;
  height: fit-content;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #606266;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.swap-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}
.swap-btn {
  transform: rotate(90deg);
  transition: all 0.3s;
  width: 44px;
  height: 44px;
  font-size: 18px;
  background-color: #ecf5ff;
  border-color: #d9ecff;
  color: #409eff;
}
.swap-btn:hover {
  transform: rotate(270deg);
  background-color: #409eff;
  border-color: #409eff;
  color: white;
}

.result-box {
  background: linear-gradient(135deg, #ecf5ff 0%, #f0f9eb 100%);
  border-radius: 12px;
  padding: 32px;
  text-align: center;
  margin: 32px 0;
  border: 1px solid #d9ecff;
  transition: all 0.3s;
}
.result-box:hover { box-shadow: 0 4px 12px rgba(64, 158, 255, 0.1); }

.result-text {
  font-size: 42px;
  font-weight: 700;
  color: #409EFF;
  line-height: 1.2;
  display: flex;
  justify-content: center;
  align-items: baseline;
}
.approx { font-size: 28px; color: #909399; font-weight: 400; margin-right: 8px; }
.currency-unit { font-size: 20px; color: #606266; margin-left: 8px; font-weight: 500; }

.result-placeholder { font-size: 20px; color: #C0C4CC; padding: 10px 0; }
.rate-hint { margin-top: 12px; font-size: 14px; color: #909399; }

.exchange-btn {
  width: 100%;
  height: 52px;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 2px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
  transition: all 0.3s;
}
.exchange-btn:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(64, 158, 255, 0.4); }

@media (max-width: 900px) {
  .card-body-layout { flex-direction: column; gap: 20px; }
  .swap-container { margin: 12px 0; }
  .swap-btn { transform: rotate(0deg); }
  .swap-btn:hover { transform: rotate(180deg); }
  .rates-panel { margin-top: 20px; }
}
</style>





  