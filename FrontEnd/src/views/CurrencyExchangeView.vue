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
            <el-tag type="info" effect="plain" size="large">数据仅供参考</el-tag>
          </div>
        </div>
      </template>

      <div class="card-body-layout">
        <div class="exchange-core">
          <el-form :model="form" label-position="top" class="exchange-form" size="large">
            <el-row :gutter="20" align="middle" class="form-row">
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

        <div class="rates-panel">
          <div class="panel-title">
            <el-icon><TrendCharts /></el-icon> 
            {{ form.fromCurrency }} 兑换其他主要货币汇率
          </div>
          <el-table 
            :data="relatedRates" size="default"
            :show-header="true"
            stripe
            style="width: 100%;"
          >
            <el-table-column prop="fromCurrency" label="基准" width="80" />
            <el-table-column prop="toCurrency" label="目标" width="80" />
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
import axios from '../axios'; // 假设你的 axios 实例已经配置好

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
    // 提取所有货币种类
    const curSet = new Set<string>();
    exchangeRates.value.forEach(r => { curSet.add(r.fromCurrency); curSet.add(r.toCurrency); });
    currencies.value = [...curSet].sort((a, b) => a.localeCompare(b));

    // 如果初始化货币不在列表中，使用默认值
    if (!currencies.value.includes(form.value.fromCurrency) && currencies.value.length > 0) {
        form.value.fromCurrency = currencies.value[0];
    }
     if (!currencies.value.includes(form.value.toCurrency) && currencies.value.length > 1) {
        form.value.toCurrency = currencies.value[1];
    }

  } catch (err) { ElMessage.error("获取汇率失败"); }
  loading.value = false;
  calculateResult();
};

const fetchLatestRate = (from: string, to: string) => {
  // 查找正向汇率
  const rateObj = exchangeRates.value.find(r => r.fromCurrency === from && r.toCurrency === to);
  if (rateObj) return rateObj.rate;
  
  // 尝试查找逆向汇率并取倒数 (例如 USD->EUR 找不到，找 EUR->USD)
  const reverseRateObj = exchangeRates.value.find(r => r.fromCurrency === to && r.toCurrency === from);
  if (reverseRateObj && reverseRateObj.rate !== 0) {
      return 1 / reverseRateObj.rate;
  }

  // 假设如果 From 和 To 相同，汇率为 1
  if (from === to) return 1;

  return null;
};

const calculateResult = () => {
  // 确保 amount 是数字且大于 0
  if (!form.value.amount || typeof form.value.amount !== 'number' || form.value.amount <= 0) { 
      result.value = null; 
      return; 
  }
  
  const rate = fetchLatestRate(form.value.fromCurrency, form.value.toCurrency);
  
  result.value = (rate !== null) ? form.value.amount * rate : null;
};

// 当前选中的 From -> To 汇率
const currentRate = computed(() => {
  const rate = fetchLatestRate(form.value.fromCurrency, form.value.toCurrency);
  return rate !== null ? rate.toFixed(4) : '---';
});

// ⭐ 修复点：新增计算属性，根据 form.fromCurrency 过滤相关汇率
const relatedRates = computed(() => {
  // 过滤出所有以当前持有货币为基准的汇率，并排除兑换到自己的情况
  const filteredRates = exchangeRates.value.filter(r => 
    r.fromCurrency === form.value.fromCurrency && r.toCurrency !== form.value.fromCurrency
  );
  
  // 补全：如果数据库中没有直接的 From->X 的汇率，可以尝试通过 From->X 的反向查找来补全。
  // 但为了简化，这里只显示直接的 From->X 汇率。
  
  return filteredRates.slice(0, 8); // 只显示前 8 条
});

const swapCurrencies = () => {
  // 使用解构赋值交换
  [form.value.fromCurrency, form.value.toCurrency] = [form.value.toCurrency, form.value.fromCurrency];
  calculateResult();
};

const exchange = () => {
  if (!form.value.amount) return ElMessage.warning("请输入金额");
  
  // 使用计算后的结果，避免重复计算和 API 调用
  if (result.value !== null) {
    ElMessage.success(`成功兑换 ${result.value.toFixed(2)} ${form.value.toCurrency}`);
  } else {
    ElMessage.error("当前货币组合无法找到有效汇率进行兑换");
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





  