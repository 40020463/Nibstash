<template>
  <div class="bookmarklet-page">
    <div class="page-header">
      <h2>Bookmarklet</h2>
    </div>

    <div class="info-card">
      <h3>什么是 Bookmarklet？</h3>
      <p>Bookmarklet 是一个特殊的书签，可以让你在浏览任何网页时，一键将当前页面添加到囤囤鼠。</p>
    </div>

    <div class="install-card">
      <h3>安装方法</h3>
      <ol>
        <li>显示浏览器的书签栏（Chrome: Ctrl+Shift+B / Mac: Cmd+Shift+B）</li>
        <li>将下面的按钮<strong>拖拽</strong>到书签栏</li>
        <li>完成！以后在任何网页点击这个书签即可收藏</li>
      </ol>

      <div class="bookmarklet-btn">
        <a
          :href="bookmarkletCode"
          class="btn"
          @click.prevent
        >
          🐿️ 收藏到囤囤鼠
        </a>
        <p class="bookmarklet-hint">← 拖拽这个按钮到书签栏</p>
      </div>

      <div class="code-section">
        <h4>手动安装</h4>
        <p>如果拖拽不起作用，可以手动创建书签：</p>
        <ol>
          <li>右键点击书签栏，选择"添加网页"或"添加书签"</li>
          <li>名称填写：<code>收藏到囤囤鼠</code></li>
          <li>网址填写下面的代码：</li>
        </ol>
        <el-input
          v-model="bookmarkletCode"
          type="textarea"
          :rows="4"
          readonly
        />
        <el-button size="small" style="margin-top: 8px" @click="copyCode">
          <el-icon><CopyDocument /></el-icon> 复制代码
        </el-button>
      </div>
    </div>

    <div class="usage-card">
      <h3>使用方法</h3>
      <ol>
        <li>在任意网页上点击书签栏中的"收藏到囤囤鼠"</li>
        <li>会弹出一个小窗口，显示添加结果</li>
        <li>如果未登录，会提示先登录</li>
      </ol>
    </div>

    <div class="usage-card">
      <h3>使用提示</h3>
      <ul>
        <li>确保已登录囤囤鼠，否则收藏时需要先登录</li>
        <li>收藏窗口会自动获取当前页面的标题和网址</li>
        <li>成功收藏后窗口会在 3 秒后自动关闭</li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ElMessage } from 'element-plus'

const baseUrl = computed(() => {
  return window.location.origin
})

const bookmarkletCode = computed(() => {
  return `javascript:(function(){var w=window.open('${baseUrl.value}/api/bookmarklet?url='+encodeURIComponent(location.href)+'&title='+encodeURIComponent(document.title),'nibstash','width=400,height=300,scrollbars=yes');w.focus();})();`
})

function copyCode() {
  navigator.clipboard.writeText(bookmarkletCode.value)
  ElMessage.success('已复制到剪贴板')
}
</script>

<style lang="scss" scoped>
.bookmarklet-page {
  max-width: 700px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;

  h2 {
    margin: 0;
  }
}

.info-card,
.install-card,
.usage-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);

  h3 {
    margin: 0 0 16px;
    font-size: 16px;
  }

  h4 {
    margin: 20px 0 12px;
    font-size: 14px;
  }

  p {
    color: #606266;
    line-height: 1.6;
  }

  ol, ul {
    padding-left: 20px;
    color: #606266;

    li {
      margin-bottom: 8px;
    }
  }

  code {
    background: #f5f7fa;
    padding: 2px 6px;
    border-radius: 4px;
    font-family: monospace;
    color: #409eff;
  }
}

.bookmarklet-btn {
  text-align: center;
  padding: 20px;

  .btn {
    display: inline-block;
    padding: 12px 24px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: #fff;
    text-decoration: none;
    border-radius: 8px;
    font-size: 16px;
    font-weight: 500;
    cursor: move;
    transition: transform 0.2s, box-shadow 0.2s;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
    }
  }

  .bookmarklet-hint {
    margin-top: 12px;
    font-size: 13px;
    color: #909399;
  }
}

.code-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
}
</style>
