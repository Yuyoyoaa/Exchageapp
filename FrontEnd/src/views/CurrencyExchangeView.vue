<template>  
  <el-container style="min-height: calc(100vh - 60px);">  
    <el-main>
      <el-card class="exchange-card">
        <template #header>
          <div class="card-header">
            <h2>货币兑换</h2>
            <el-button type="primary" @click="fetchExchangeRates" :loading="loading">
              刷新汇率
            </el-button>
          </div>
        </template>

        <el-form :model="form" class="exchange-form">
          <el-row :gutter="20">
            <el-col :span="10">
              <el-form-item label="从货币" label-width="80px">
                <el-select v-model="form.fromCurrency" placeholder="选择货币" style="width: 100%;">
                  <el-option v-for="c in currencies" :key="c" :value="c" :label="c" />
                </el-select>
              </el-form-item>
            </el-col>

            <el-col :span="10">
              <el-form-item label="到货币" label-width="80px">
                <el-select v-model="form.toCurrency" placeholder="选择货币" style="width: 100%;">
                  <el-option v-for="c in currencies" :key="c" :value="c" :label="c" />
                </el-select>
              </el-form-item>
            </el-col>

            <el-col :span="4">
              <el-button type="success" style="width:100%;margin-top:32px" @click="swapCurrencies">
                ⇄ 交换
              </el-button>
            </el-col>
          </el-row>

          <el-form-item label="金额" label-width="80px">
            <el-input v-model="form.amount" type="number" @input="calculateResult">
              <template #prepend>{{ form.fromCurrency }}</template>
            </el-input>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" style="width:100%" @click="exchange">立即兑换</el-button>
          </el-form-item>
        </el-form>

        <!-- 结果展示 -->
        <div v-if="result !== null" class="result-panel">
          <el-alert
            title="兑换结果"
            type="success"
            :description="`${form.amount} ${form.fromCurrency} = ${result.toFixed(2)} ${form.toCurrency}`"
            :closable="false"
            show-icon
          />
          <p class="rate-info">
            当前汇率：1 {{ form.fromCurrency }} = {{ currentRate }} {{ form.toCurrency }}
          </p>
        </div>

        <!-- 汇率表 -->
        <div class="rates-table" v-if="exchangeRates.length">
          <h3>最新汇率</h3>
          <el-table :data="exchangeRates">
            <el-table-column prop="fromCurrency" label="从货币" width="120" />
            <el-table-column prop="toCurrency" label="到货币" width="120" />
            <el-table-column prop="rate" label="汇率" />
            <el-table-column label="日期">
              <template #default="s">
                {{ formatDate(s.row.date) }}
              </template>
            </el-table-column>
          </el-table>
        </div>

      </el-card>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
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

const form = ref({
  fromCurrency: 'USD',
  toCurrency: 'CNY',
  amount: 100
});

const exchangeRates = ref<ExchangeRate[]>([]);
const currencies = ref<string[]>([]);
const result = ref<number | null>(null);
const loading = ref(false);

const fetchExchangeRates = async () => {
  loading.value = true;
  try {
    const res = await axios.get('/exchangeRates');
    exchangeRates.value = res.data;

    const cur = new Set<string>();
    exchangeRates.value.forEach(r => {
      cur.add(r.fromCurrency);
      cur.add(r.toCurrency);
    });
    currencies.value = [...cur];
  } catch {
    ElMessage.error("获取汇率失败");
  }
  loading.value = false;
};

const calculateResult = () => {
  const rate = exchangeRates.value.find(
    r => r.fromCurrency === form.value.fromCurrency && r.toCurrency === form.value.toCurrency
  );
  result.value = rate ? form.value.amount * rate.rate : null;
};

const currentRate = computed(() => {
  const rate = exchangeRates.value.find(
    r => r.fromCurrency === form.value.fromCurrency && r.toCurrency === form.value.toCurrency
  );
  return rate ? rate.rate : 'N/A';
});

const swapCurrencies = () => {
  [form.value.fromCurrency, form.value.toCurrency] =
    [form.value.toCurrency, form.value.fromCurrency];
  calculateResult();
};

const exchange = () => {
  ElMessage.success("兑换成功（演示，无后端功能）");
};

const formatDate = (d: string) => new Date(d).toLocaleString("zh-CN");

onMounted(fetchExchangeRates);
</script>

  