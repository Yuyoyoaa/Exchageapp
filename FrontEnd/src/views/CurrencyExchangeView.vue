<template>
  <div class="exchange-wrapper">
    <el-card class="exchange-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <el-icon class="icon-money" size="28" color="#409EFF"><Money /></el-icon>
            <h2>货币兑换中心</h2>
          </div>
          <el-button 
            type="primary" 
            plain 
            round
            @click="fetchExchangeRates" 
            :loading="loading"
            icon="Refresh"
          >
            刷新汇率
          </el-button>
        </div>
      </template>

      <div class="card-body-layout">
        <!-- 左侧：操作区 -->
        <div class="exchange-core">
          <el-form :model="form" label-position="top" class="exchange-form" size="large">
            <el-row :gutter="20" align="middle" class="form-row">
              <!-- 从货币 -->
              <el-col :xs="24" :sm="10">
                <el-form-item label="持有货币">
                  <el-select 
                    v-model="form.fromCurrency" 
                    filterable
                    style="width: 100%;" 
                    @change="calculateResult"
                  >
                    <el-option v-for="c in currencies" :key="c" :value="c" :label="c" />
                  </el-select>
                </el-form-item>
              </el-col>

              <!-- 交换按钮 -->
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

              <!-- 到货币 -->
              <el-col :xs="24" :sm="10">
                <el-form-item label="目标货币">
                  <el-select 
                    v-model="form.toCurrency" 
                    filterable
                    style="width: 100%;" 
                    @change="calculateResult"
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

            <!-- 结果展示区 -->
            <div class="result-box">
              <div v-if="result !== null" class="result-text">
                <span class="approx">≈</span> {{ result.toFixed(2) }} <span class="currency-unit">{{ form.toCurrency }}</span>
              </div>
              <div v-else class="result-placeholder">请输入金额查看结果</div>
              <div class="rate-hint">
                参考汇率：1 {{ form.fromCurrency }} ≈ {{ currentRate }} {{ form.toCurrency }}
              </div>
            </div>

            <el-button type="primary" class="exchange-btn" @click="exchange">
              立即兑换
            </el-button>
          </el-form>
        </div>

        <!-- 右侧：行情表（在大屏时分栏显示） -->
        <div class="rates-panel">
          <div class="panel-title">
            <el-icon><TrendCharts /></el-icon> 相关汇率行情
          </div>
          <el-table 
            :data="exchangeRates.slice(0, 8)" 
            size="default"
            :show-header="true"
            stripe
            style="width: 100%;"
          >
            <el-table-column prop="fromCurrency" label="基准" width="80" />
            <el-table-column prop="toCurrency" label="目标" width="80" />
            <el-table-column prop="rate" label="汇率" align="right">
                <template #default="scope">
                  <span style="color: #E6A23C; font-weight: bold;">{{ scope.row.rate }}</span>
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
  id: number;
  fromCurrency: string;
  toCurrency: string;
  rate: number;
  date: string;
}

const form = ref({ fromCurrency: 'USD', toCurrency: 'CNY', amount: 100 });
const exchangeRates = ref<ExchangeRate[]>([]);
const currencies = ref<string[]>([]);
const result = ref<number | null>(null);
const loading = ref(false);

const fetchExchangeRates = async () => {
  loading.value = true;
  try {
    const res = await axios.get('/exchangeRates');
    exchangeRates.value = res.data;
    const curSet = new Set<string>();
    exchangeRates.value.forEach(r => { curSet.add(r.fromCurrency); curSet.add(r.toCurrency); });
    currencies.value = [...curSet].sort((a, b) => a.localeCompare(b));
  } catch (err) { ElMessage.error("获取汇率失败"); }
  loading.value = false;
  calculateResult();
};

const fetchLatestRate = async (from: string, to: string) => {
  const rateObj = exchangeRates.value.find(r => r.fromCurrency === from && r.toCurrency === to);
  return rateObj ? rateObj.rate : null;
};

const calculateResult = async () => {
  if (!form.value.amount) { result.value = null; return; }
  const rate = await fetchLatestRate(form.value.fromCurrency, form.value.toCurrency);
  result.value = (rate !== null) ? form.value.amount * rate : null;
};

const currentRate = computed(() => {
  const rate = exchangeRates.value.find(r => r.fromCurrency === form.value.fromCurrency && r.toCurrency === form.value.toCurrency);
  return rate ? rate.rate : '---';
});

const swapCurrencies = () => {
  [form.value.fromCurrency, form.value.toCurrency] = [form.value.toCurrency, form.value.fromCurrency];
  calculateResult();
};

const exchange = async () => {
  if (!form.value.amount) return ElMessage.warning("请输入金额");
  const rate = await fetchLatestRate(form.value.fromCurrency, form.value.toCurrency);
  if (rate !== null) {
    result.value = form.value.amount * rate;
    ElMessage.success(`兑换成功！`);
  } else {
    ElMessage.error("无法兑换");
  }
};

onMounted(fetchExchangeRates);
</script>

<style scoped>
.exchange-wrapper {
  min-height: calc(100vh - 60px);
  background: #f5f7fa;
  /* 使用 Flex 居中，保留一点内边距 */
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  box-sizing: border-box;
}

/* 核心修改：让卡片变大 */
.exchange-card {
  width: 100%;
  max-width: 1200px; /* 增加最大宽度，适配宽屏 */
  min-height: 600px; /* 增加最小高度 */
  border-radius: 12px;
  display: flex;
  flex-direction: column;
}

/* 头部样式 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
}
.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #303133;
}
.header-title h2 { margin: 0; font-size: 22px; }

/* 布局核心：左右分栏 */
.card-body-layout {
  display: flex;
  gap: 40px;
  padding: 20px 0;
  height: 100%;
}

/* 左侧操作区：占据主要空间 */
.exchange-core {
  flex: 1.5; /* 左侧更宽 */
  display: flex;
  flex-direction: column;
  justify-content: center;
}

/* 右侧行情区 */
.rates-panel {
  flex: 1;
  background: #fcfcfc;
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #ebeef5;
  display: flex;
  flex-direction: column;
}

.panel-title {
  font-size: 16px;
  font-weight: bold;
  color: #606266;
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 表单元素优化 */
.swap-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}
.swap-btn {
  transform: rotate(90deg);
  transition: transform 0.3s;
  width: 48px;
  height: 48px;
  font-size: 20px;
}
.swap-btn:hover { transform: rotate(270deg); }

/* 结果展示区优化 */
.result-box {
  background: #ecf5ff;
  border-radius: 8px;
  padding: 30px;
  text-align: center;
  margin: 30px 0;
  border: 1px solid #d9ecff;
}

.result-text {
  font-size: 48px; /* 字体加大 */
  font-weight: bold;
  color: #409EFF;
  line-height: 1.2;
}
.approx { font-size: 32px; color: #909399; font-weight: normal; margin-right: 5px; }
.result-placeholder {
  font-size: 24px;
  color: #C0C4CC;
}
.currency-unit { font-size: 24px; color: #606266; margin-left: 8px; }
.rate-hint { margin-top: 10px; font-size: 14px; color: #909399; }

.exchange-btn {
  width: 100%;
  height: 50px;
  font-size: 18px;
  font-weight: bold;
  letter-spacing: 2px;
}

/* 响应式：小屏变回单列 */
@media (max-width: 900px) {
  .card-body-layout {
    flex-direction: column;
  }
  .swap-container { margin: 10px 0; }
  .swap-btn { transform: rotate(0deg); }
  .swap-btn:hover { transform: rotate(180deg); }
  
  .exchange-card {
    max-width: 100%;
  }
}
</style>





  