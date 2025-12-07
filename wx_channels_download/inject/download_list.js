// 悬浮下载列表组件
var __wx_channels_download_list__ = {
  list: [], // 下载列表数据
  container: null, // 列表容器
  isExpanded: false, // 是否展开
  maxItems: 10, // 最大显示数量
};

// 初始化下载列表
function init_download_list() {
  if (document.getElementById("__wx_channels_download_list__")) {
    return;
  }

  // 创建主容器
  var container = document.createElement("div");
  container.id = "__wx_channels_download_list__";
  container.style.cssText =
    "position: fixed; right: 24px; top: 160px; z-index: 999998; " +
    "background: #fff; border-radius: 12px; box-shadow: 0 4px 16px rgba(0,0,0,.15); " +
    "min-width: 320px; max-width: 400px; max-height: 500px; " +
    "overflow: hidden; display: none; transition: all 0.3s ease;";

  // 创建标题栏
  var header = document.createElement("div");
  header.style.cssText =
    "padding: 12px 16px; border-bottom: 1px solid #eee; " +
    "display: flex; justify-content: space-between; align-items: center; " +
    "background: #f7f7f7; cursor: pointer;";
  header.innerHTML = `
    <div style="font-weight: 600; font-size: 14px; color: #333;">
      下载列表 <span id="__wx_download_count__" style="color: #999; font-weight: normal;">(0)</span>
    </div>
    <div style="display: flex; flex-direction: column; align-items: flex-end; gap: 4px;">
      <!-- 积分显示区域 -->
      <div id="__wx_credit_info__" style="font-size: 11px; display: flex; align-items: center; gap: 6px;">
        <span style="color: #666;">积分:</span>
        <span id="__wx_credit_points__" style="color: #07c160; font-weight: 600;">--</span>
        <span id="__wx_credit_expires__" style="color: #999; font-size: 10px;">(--)</span>
      </div>
      <!-- 原有按钮 -->
      <div style="display: flex; gap: 8px; align-items: center;">
        <span id="__wx_download_mp3__" style="display: none; font-size: 11px; color: #1890ff; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: background 0.2s;">下载MP3</span>
        <span id="__wx_download_cover__" style="font-size: 11px; color: #1890ff; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: background 0.2s;">下载封面</span>
        <span id="__wx_download_toggle__" style="font-size: 12px; color: #666;">▼</span>
        <span id="__wx_download_clear__" style="font-size: 12px; color: #07c160; cursor: pointer;">清空</span>
      </div>
    </div>
  `;

  // 创建列表内容区域
  var listContent = document.createElement("div");
  listContent.id = "__wx_download_list_content__";
  listContent.style.cssText =
    "max-height: 400px; overflow-y: auto; " +
    "scrollbar-width: thin; scrollbar-color: #ccc transparent;";

  // 空状态提示
  var emptyState = document.createElement("div");
  emptyState.id = "__wx_download_empty__";
  emptyState.style.cssText =
    "padding: 40px 20px; text-align: center; color: #999; font-size: 13px;";
  emptyState.textContent = "暂无下载记录";

  listContent.appendChild(emptyState);
  container.appendChild(header);
  container.appendChild(listContent);
  document.body.appendChild(container);

  __wx_channels_download_list__.container = container;

  // 绑定事件
  header.onclick = function (e) {
    if (e.target.id === "__wx_download_clear__") {
      e.stopPropagation();
      clear_download_list();
      return;
    }
    if (e.target.id === "__wx_download_mp3__") {
      e.stopPropagation();
      var profile = window.__wx_channels_store__.profile;
      if (!profile) {
        if (window.__wx_channels_tip__ && window.__wx_channels_tip__.toast) {
          window.__wx_channels_tip__.toast("没有视频数据", 2000);
        }
        return;
      }
      var filename = __wx_build_filename(profile, null, __wx_channels_config__.downloadFilenameTemplate);
      if (filename && typeof __wx_channels_download_as_mp3 === "function") {
        __wx_channels_download_as_mp3(profile, filename);
      }
      return;
    }
    if (e.target.id === "__wx_download_cover__") {
      e.stopPropagation();
      if (typeof __wx_channels_handle_download_cover === "function") {
        __wx_channels_handle_download_cover();
      }
      return;
    }
    toggle_download_list();
  };

  // 添加按钮悬停效果
  var mp3Btn = document.getElementById("__wx_download_mp3__");
  if (mp3Btn) {
    mp3Btn.onmouseenter = function() {
      this.style.background = "#1890ff15";
    };
    mp3Btn.onmouseleave = function() {
      this.style.background = "transparent";
    };
  }
  var coverBtn = document.getElementById("__wx_download_cover__");
  if (coverBtn) {
    coverBtn.onmouseenter = function() {
      this.style.background = "#1890ff15";
    };
    coverBtn.onmouseleave = function() {
      this.style.background = "transparent";
    };
  }

  // 点击外部关闭
  document.addEventListener("click", function (e) {
    if (
      container.contains(e.target) ||
      e.target.id === "__wx_channels_floating_download_btn__"
    ) {
      return;
    }
    if (__wx_channels_download_list__.isExpanded) {
      collapse_download_list();
    }
  });
}

// 检查是否已存在相同的下载项
function find_existing_download_item(profile, spec) {
  if (!profile || !profile.id) {
    return null;
  }
  
  var specFormat = spec ? spec.fileFormat : 'original';
  
  // 查找相同视频ID和规格的项
  return __wx_channels_download_list__.list.find(function(item) {
    if (!item.profile || !item.profile.id) {
      return false;
    }
    
    var itemSpecFormat = item.spec ? item.spec.fileFormat : 'original';
    
    // 匹配视频ID和规格
    return item.profile.id === profile.id && itemSpecFormat === specFormat;
  });
}

// 添加下载项到列表
function add_to_download_list(profile, spec, status, filename) {
  if (!__wx_channels_download_list__.container) {
    init_download_list();
  }

  // 检查是否已存在相同的下载项
  var existingItem = find_existing_download_item(profile, spec);
  
  if (existingItem) {
    // 如果已存在且正在下载，不重复添加
    if (existingItem.status === "downloading") {
      if (window.__wx_channels_tip__ && window.__wx_channels_tip__.toast) {
        window.__wx_channels_tip__.toast("该视频正在下载中", 2000);
      }
      return existingItem.id;
    }
    
    // 如果已存在但已完成或失败，更新状态为下载中并更新时间戳
    existingItem.status = status;
    existingItem.timestamp = Date.now();
    existingItem.filename = filename || __wx_build_filename(profile, spec, __wx_channels_config__.downloadFilenameTemplate);
    
    // 将该项移到列表开头
    var index = __wx_channels_download_list__.list.indexOf(existingItem);
    if (index > 0) {
      __wx_channels_download_list__.list.splice(index, 1);
      __wx_channels_download_list__.list.unshift(existingItem);
    }
    
    update_download_list_display();
    show_download_list();
    return existingItem.id;
  }

  // 创建新项
  var item = {
    id: __wx_uid__(),
    profile: profile,
    spec: spec,
    status: status, // 'downloading', 'completed', 'failed'
    filename: filename || __wx_build_filename(profile, spec, __wx_channels_config__.downloadFilenameTemplate),
    timestamp: Date.now(),
    url: profile.url + (spec ? "&X-snsvideoflag=" + spec.fileFormat : ""),
    key: profile.key,
    progress: 0, // 下载进度百分比 (0-100)
  };

  // 添加到列表开头
  __wx_channels_download_list__.list.unshift(item);

  // 限制列表长度
  if (__wx_channels_download_list__.list.length > __wx_channels_download_list__.maxItems) {
    __wx_channels_download_list__.list = __wx_channels_download_list__.list.slice(
      0,
      __wx_channels_download_list__.maxItems
    );
  }

  update_download_list_display();
  show_download_list();
  return item.id;
}

// 更新下载项状态
function update_download_item_status(id, status, error) {
  var item = __wx_channels_download_list__.list.find((i) => i.id === id);
  if (item) {
    item.status = status;
    if (error) {
      item.error = error;
    }
    // 如果状态变为已完成或失败，重置进度
    if (status === "completed" || status === "failed") {
      item.progress = status === "completed" ? 100 : 0;
    }
    update_download_list_display();
  }
}

// 更新下载项进度
function update_download_item_progress(id, progress) {
  var item = __wx_channels_download_list__.list.find((i) => i.id === id);
  if (item && item.status === "downloading") {
    item.progress = Math.min(100, Math.max(0, progress));
    // 立即更新显示
    update_download_list_display();
  }
}

// 将函数暴露到全局，以便 main.js 可以调用
window.update_download_item_progress = update_download_item_progress;

// 更新列表显示
function update_download_list_display() {
  var listContent = document.getElementById("__wx_download_list_content__");
  var emptyState = document.getElementById("__wx_download_empty__");
  var countEl = document.getElementById("__wx_download_count__");

  if (!listContent) return;

  // 更新计数
  if (countEl) {
    var total = __wx_channels_download_list__.list.length;
    var completed = __wx_channels_download_list__.list.filter(
      (i) => i.status === "completed"
    ).length;
    countEl.textContent = `(${completed}/${total})`;
  }

  // 清空内容
  listContent.innerHTML = "";

  if (__wx_channels_download_list__.list.length === 0) {
    if (emptyState) {
      listContent.appendChild(emptyState);
    }
    return;
  }

  // 渲染列表项
  __wx_channels_download_list__.list.forEach(function (item) {
    var listItem = create_download_list_item(item);
    listContent.appendChild(listItem);
  });
}

// 创建列表项元素
function create_download_list_item(item) {
  var itemEl = document.createElement("div");
  itemEl.style.cssText =
    "padding: 12px 16px; border-bottom: 1px solid #f0f0f0; " +
    "display: flex; flex-direction: column; gap: 8px; " +
    "transition: background 0.2s; cursor: pointer;";
  itemEl.onmouseenter = function () {
    this.style.background = "#f7f7f7";
  };
  itemEl.onmouseleave = function () {
    this.style.background = "transparent";
  };

  // 标题和状态
  var header = document.createElement("div");
  header.style.cssText = "display: flex; justify-content: space-between; align-items: center;";

  var title = document.createElement("div");
  title.style.cssText =
    "font-size: 13px; font-weight: 500; color: #333; " +
    "overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1;";
  title.textContent = item.filename || item.profile.title || "未命名视频";
  title.title = item.filename || item.profile.title || "未命名视频";

  var statusBadge = document.createElement("span");
  var statusConfig = {
    downloading: { text: "下载中", color: "#07c160" },
    completed: { text: "已完成", color: "#1890ff" },
    failed: { text: "失败", color: "#ff4d4f" },
  };
  var config = statusConfig[item.status] || statusConfig.downloading;
  statusBadge.style.cssText =
    "font-size: 11px; padding: 2px 8px; border-radius: 10px; " +
    "background: " +
    config.color +
    "15; color: " +
    config.color +
    "; white-space: nowrap;";
  statusBadge.textContent = config.text;

  header.appendChild(title);
  header.appendChild(statusBadge);

  // 详细信息
  var info = document.createElement("div");
  info.style.cssText = "font-size: 11px; color: #999; display: flex; gap: 12px;";

  var specText = item.spec ? item.spec.fileFormat : "原始";
  var timeText = new Date(item.timestamp).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  });

  info.innerHTML = `
    <span>${specText}</span>
    <span>${timeText}</span>
  `;

  // 进度条（仅在下载中时显示）
  if (item.status === "downloading") {
    var progressContainer = document.createElement("div");
    progressContainer.style.cssText = "margin-top: 6px;";

    var progressBarBg = document.createElement("div");
    progressBarBg.style.cssText =
      "width: 100%; height: 4px; background: #f0f0f0; border-radius: 2px; overflow: hidden;";

    var progressValue = item.progress !== undefined && item.progress !== null ? item.progress : 0;
    var progressBarFill = document.createElement("div");
    progressBarFill.style.cssText =
      "height: 100%; background: linear-gradient(90deg, #07c160 0%, #52c41a 100%); " +
      "border-radius: 2px; transition: width 0.3s ease; " +
      "width: " + progressValue + "%;";

    progressBarBg.appendChild(progressBarFill);
    progressContainer.appendChild(progressBarBg);

    var progressText = document.createElement("div");
    progressText.style.cssText =
      "font-size: 11px; color: #07c160; margin-top: 4px; text-align: right;";
    progressText.textContent = progressValue.toFixed(1) + "%";

    progressContainer.appendChild(progressText);
    itemEl.appendChild(progressContainer);
  }

  // 操作按钮
  var actions = document.createElement("div");
  actions.style.cssText =
    "display: flex; gap: 8px; margin-top: 4px; " +
    "padding-top: 8px; border-top: 1px solid #f0f0f0;";

  if (item.status === "completed") {
    // 重新下载按钮
    var redownloadBtn = document.createElement("span");
    redownloadBtn.style.cssText =
      "font-size: 11px; color: #07c160; cursor: pointer; " +
      "padding: 4px 8px; border-radius: 4px; " +
      "transition: background 0.2s;";
    redownloadBtn.textContent = "重新下载";
    redownloadBtn.onmouseenter = function () {
      this.style.background = "#07c16015";
    };
    redownloadBtn.onmouseleave = function () {
      this.style.background = "transparent";
    };
    redownloadBtn.onclick = function (e) {
      e.stopPropagation();
      __wx_channels_handle_click_download__(item.spec);
    };

    // 复制命令按钮
    var copyCmdBtn = document.createElement("span");
    copyCmdBtn.style.cssText =
      "font-size: 11px; color: #1890ff; cursor: pointer; " +
      "padding: 4px 8px; border-radius: 4px; " +
      "transition: background 0.2s;";
    copyCmdBtn.textContent = "复制命令";
    copyCmdBtn.onmouseenter = function () {
      this.style.background = "#1890ff15";
    };
    copyCmdBtn.onmouseleave = function () {
      this.style.background = "transparent";
    };
    copyCmdBtn.onclick = function (e) {
      e.stopPropagation();
      var command = `download --url "${item.url}"`;
      if (item.key) {
        command += ` --key ${item.key}`;
      }
      command += ` --filename "${item.filename}.mp4"`;
      __wx_channels_copy(command);
      if (window.__wx_channels_tip__ && window.__wx_channels_tip__.toast) {
        window.__wx_channels_tip__.toast("命令已复制", 1000);
      }
    };

    actions.appendChild(redownloadBtn);
    actions.appendChild(copyCmdBtn);
  } else if (item.status === "failed") {
    // 重试按钮
    var retryBtn = document.createElement("span");
    retryBtn.style.cssText =
      "font-size: 11px; color: #ff4d4f; cursor: pointer; " +
      "padding: 4px 8px; border-radius: 4px; " +
      "transition: background 0.2s;";
    retryBtn.textContent = "重试";
    retryBtn.onmouseenter = function () {
      this.style.background = "#ff4d4f15";
    };
    retryBtn.onmouseleave = function () {
      this.style.background = "transparent";
    };
    retryBtn.onclick = function (e) {
      e.stopPropagation();
      update_download_item_status(item.id, "downloading");
      __wx_channels_handle_click_download__(item.spec);
    };
    actions.appendChild(retryBtn);
  }

  // 删除按钮
  var deleteBtn = document.createElement("span");
  deleteBtn.style.cssText =
    "font-size: 11px; color: #999; cursor: pointer; " +
    "padding: 4px 8px; border-radius: 4px; " +
    "transition: background 0.2s; margin-left: auto;";
  deleteBtn.textContent = "删除";
  deleteBtn.onmouseenter = function () {
    this.style.background = "#f0f0f0";
  };
  deleteBtn.onmouseleave = function () {
    this.style.background = "transparent";
  };
  deleteBtn.onclick = function (e) {
    e.stopPropagation();
    remove_from_download_list(item.id);
  };
  actions.appendChild(deleteBtn);

  itemEl.appendChild(header);
  itemEl.appendChild(info);
  if (actions.children.length > 0) {
    itemEl.appendChild(actions);
  }

  return itemEl;
}

// 从列表移除项
function remove_from_download_list(id) {
  __wx_channels_download_list__.list = __wx_channels_download_list__.list.filter(
    (i) => i.id !== id
  );
  update_download_list_display();
  if (__wx_channels_download_list__.list.length === 0) {
    hide_download_list();
  }
}

// 清空下载列表
function clear_download_list() {
  if (confirm("确定要清空所有下载记录吗？")) {
    __wx_channels_download_list__.list = [];
    update_download_list_display();
    hide_download_list();
  }
}

// 显示下载列表
function show_download_list() {
  if (!__wx_channels_download_list__.container) return;
  __wx_channels_download_list__.container.style.display = "block";
  if (!__wx_channels_download_list__.isExpanded) {
    expand_download_list();
  }
}

// 隐藏下载列表
function hide_download_list() {
  if (!__wx_channels_download_list__.container) return;
  __wx_channels_download_list__.container.style.display = "none";
}

// 展开列表
function expand_download_list() {
  __wx_channels_download_list__.isExpanded = true;
  var toggle = document.getElementById("__wx_download_toggle__");
  if (toggle) {
    toggle.textContent = "▲";
  }
  var listContent = document.getElementById("__wx_download_list_content__");
  if (listContent) {
    listContent.style.display = "block";
  }
}

// 折叠列表
function collapse_download_list() {
  __wx_channels_download_list__.isExpanded = false;
  var toggle = document.getElementById("__wx_download_toggle__");
  if (toggle) {
    toggle.textContent = "▼";
  }
  var listContent = document.getElementById("__wx_download_list_content__");
  if (listContent) {
    listContent.style.display = "none";
  }
}

// 切换列表展开/折叠
function toggle_download_list() {
  if (__wx_channels_download_list__.isExpanded) {
    collapse_download_list();
  } else {
    expand_download_list();
  }
}

// 修改悬浮下载按钮，添加点击显示列表的功能
function modify_floating_download_btn() {
  var btnContainer = document.getElementById("__wx_channels_floating_download_btn__");
  if (!btnContainer) return;
  
  var btn = btnContainer.querySelector("div:first-child");
  if (!btn) return;

  // 添加右键菜单或长按显示列表
  var longPressTimer = null;
  btn.onmousedown = function (e) {
    if (e.button === 2) {
      // 右键
      e.preventDefault();
      show_download_list();
      return;
    }
    // 长按
    longPressTimer = setTimeout(function () {
      show_download_list();
    }, 500);
  };

  btn.onmouseup = function () {
    if (longPressTimer) {
      clearTimeout(longPressTimer);
      longPressTimer = null;
    }
  };

  btn.oncontextmenu = function (e) {
    e.preventDefault();
    show_download_list();
  };

  // 添加列表图标按钮
  var listIcon = document.createElement("div");
  listIcon.style.cssText =
    "width: 40px; height: 40px; background: #fff; border-radius: 50%; " +
    "display: flex; align-items: center; justify-content: center; " +
    "box-shadow: 0 2px 8px rgba(0,0,0,.15); cursor: pointer; " +
    "font-size: 18px; transition: all 0.3s ease; " +
    "border: 2px solid #07c160;";
  listIcon.innerHTML = "📋";
  listIcon.title = "查看下载列表（右键或长按主按钮）";
  
  listIcon.onmouseenter = function() {
    this.style.transform = "scale(1.1)";
    this.style.boxShadow = "0 4px 12px rgba(0,0,0,.2)";
  };
  listIcon.onmouseleave = function() {
    this.style.transform = "scale(1)";
    this.style.boxShadow = "0 2px 8px rgba(0,0,0,.15)";
  };
  
  listIcon.onclick = function (e) {
    e.stopPropagation();
    show_download_list();
  };
  
  btnContainer.appendChild(listIcon);
}

// 包装下载函数，自动添加到列表
(function() {
  // 等待 main.js 加载完成后再包装
  setTimeout(function() {
    if (typeof window.__wx_channels_handle_click_download__ === 'function') {
      var original_download_handler = window.__wx_channels_handle_click_download__;
      
      window.__wx_channels_handle_click_download__ = async function (spec, mp3) {
        var profile = __wx_channels_store__.profile;
        if (!profile) {
          return original_download_handler.call(this, spec, mp3);
        }

        // 检查积分（解耦：通过 API 检查，不直接依赖积分模块）
        if (typeof window.fetch_credit_info === "function") {
          var creditCheck = await window.fetch_credit_info();
          if (!creditCheck.valid) {
            alert(creditCheck.error || "积分不足或已过期");
            return;
          }
          
          // 显示积分信息并确认
          var expiresDate = new Date(creditCheck.expires_at * 1000);
          var expiresStr = expiresDate.toLocaleDateString("zh-CN");
          if (!confirm("当前积分：" + creditCheck.points + "\n到期时间：" + expiresStr + "\n本次下载将消耗 5 积分，确认下载？")) {
            return;
          }
          
          // 消耗积分（下载视频消耗5积分）
          try {
            var consumeResponse = await fetch("/__wx_channels_api/credit/consume", {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({ cost: 5 }),
            });
            var consumeResult = await consumeResponse.json();
            if (!consumeResult.success) {
              alert(consumeResult.error || "扣除积分失败");
              return;
            }
            
            // 更新积分显示
            if (typeof window.update_credit_display === "function") {
              window.update_credit_display({
                valid: true,
                points: consumeResult.points,
                start_at: consumeResult.start_at,
                end_at: consumeResult.end_at,
                expires_at: consumeResult.expires_at // 兼容旧格式
              });
            }
          } catch (err) {
            alert("扣除积分失败: " + err.message);
            return;
          }
        }

        // 确保使用最新的 profile 数据（深拷贝避免引用问题）
        var currentProfile = JSON.parse(JSON.stringify(profile));
        
        // 添加到下载列表（会自动去重）
        var itemId = add_to_download_list(currentProfile, spec, "downloading");
        var item = __wx_channels_download_list__.list.find((i) => i.id === itemId);
        
        if (!item) {
          return original_download_handler.call(this, spec, mp3);
        }

        // 将下载项ID设置到 __wx_channels_store__.profile 上，以便下载函数更新进度
        // 因为 __wx_channels_handle_click_download__ 函数内部是从 __wx_channels_store__.profile 获取 profile
        if (__wx_channels_store__ && __wx_channels_store__.profile) {
          __wx_channels_store__.profile.downloadItemId = itemId;
        }

        // 执行下载
        try {
          var result = original_download_handler.call(this, spec, mp3);
          
          // 如果是 Promise，监听完成
          if (result && typeof result.then === 'function') {
            result.then(function() {
              if (item) {
                update_download_item_status(item.id, "completed");
                // 更新积分显示（如果可用）
                if (typeof window.fetch_credit_info === "function") {
                  window.fetch_credit_info().then(function(creditInfo) {
                    if (typeof window.update_credit_display === "function") {
                      window.update_credit_display(creditInfo);
                    }
                  });
                }
              }
            }).catch(function(err) {
              if (item) {
                update_download_item_status(item.id, "failed", err.message || String(err));
              }
            });
          } else {
            // 非异步，延迟标记为完成（实际下载可能还在进行）
            setTimeout(function () {
              if (item && item.status === "downloading") {
                update_download_item_status(item.id, "completed");
              }
            }, 3000);
          }
          return result;
        } catch (err) {
          if (item) {
            update_download_item_status(item.id, "failed", err.message || String(err));
          }
          throw err;
        }
      };
    }
  }, 500);
})();

// 更新积分显示
function update_credit_display(creditInfo) {
  var pointsEl = document.getElementById("__wx_credit_points__");
  var expiresEl = document.getElementById("__wx_credit_expires__");
  var creditInfoEl = document.getElementById("__wx_credit_info__");
  
  if (!pointsEl || !expiresEl) {
    return;
  }
  
  if (!creditInfo || !creditInfo.valid) {
    pointsEl.textContent = "0";
    expiresEl.textContent = creditInfo?.error || "未配置";
    if (creditInfoEl) {
      creditInfoEl.style.color = "#ff4d4f";
    }
    return;
  }
  
  // 更新积分数量
  var points = creditInfo.points || 0;
  pointsEl.textContent = points;
  pointsEl.style.color = points < 5 ? "#ff4d4f" : "#07c160";
  
  // 更新到期时间（显示日期区间）
  if (creditInfo.start_at && creditInfo.end_at) {
    var startDate = new Date(creditInfo.start_at * 1000);
    var endDate = new Date(creditInfo.end_at * 1000);
    var now = new Date();
    
    if (now < startDate) {
      // 尚未生效
      expiresEl.textContent = "(" + startDate.toLocaleDateString("zh-CN") + "生效)";
      expiresEl.style.color = "#1890ff";
      if (creditInfoEl) {
        creditInfoEl.style.color = "#666";
      }
    } else if (now > endDate) {
      // 已过期
      expiresEl.textContent = "(已过期)";
      expiresEl.style.color = "#ff4d4f";
      if (creditInfoEl) {
        creditInfoEl.style.color = "#ff4d4f";
      }
    } else {
      // 有效期内，显示结束日期
      var daysLeft = Math.ceil((endDate - now) / (1000 * 60 * 60 * 24));
      if (daysLeft <= 3) {
        expiresEl.textContent = "(" + daysLeft + "天后过期)";
        expiresEl.style.color = "#ff9800";
      } else {
        expiresEl.textContent = "(" + startDate.toLocaleDateString("zh-CN") + " ~ " + endDate.toLocaleDateString("zh-CN") + ")";
        expiresEl.style.color = "#999";
      }
      if (creditInfoEl) {
        creditInfoEl.style.color = "#666";
      }
    }
  } else if (creditInfo.expires_at) {
    // 兼容旧格式（如果存在）
    var expiresDate = new Date(creditInfo.expires_at * 1000);
    var now = new Date();
    var daysLeft = Math.ceil((expiresDate - now) / (1000 * 60 * 60 * 24));
    
    if (daysLeft < 0) {
      expiresEl.textContent = "(已过期)";
      expiresEl.style.color = "#ff4d4f";
      if (creditInfoEl) {
        creditInfoEl.style.color = "#ff4d4f";
      }
    } else if (daysLeft <= 3) {
      expiresEl.textContent = "(" + daysLeft + "天后过期)";
      expiresEl.style.color = "#ff9800";
      if (creditInfoEl) {
        creditInfoEl.style.color = "#666";
      }
    } else {
      expiresEl.textContent = "(" + expiresDate.toLocaleDateString("zh-CN") + ")";
      expiresEl.style.color = "#999";
      if (creditInfoEl) {
        creditInfoEl.style.color = "#666";
      }
    }
  }
}

// 获取积分信息
async function fetch_credit_info() {
  try {
    const response = await fetch("/__wx_channels_api/credit/check", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });
    const data = await response.json();
    return data;
  } catch (err) {
    return { valid: false, error: "获取积分信息失败" };
  }
}

// 定期更新积分显示（每30秒）
function start_credit_timer() {
  // 立即更新一次
  fetch_credit_info().then(update_credit_display);
  
  // 每30秒更新一次
  setInterval(function() {
    fetch_credit_info().then(update_credit_display);
  }, 30000);
}

// 将函数暴露到全局，以便其他模块调用
window.update_credit_display = update_credit_display;
window.fetch_credit_info = fetch_credit_info;

// 初始化
setTimeout(function () {
  init_download_list();
  modify_floating_download_btn();
  // 启动积分更新定时器
  setTimeout(function() {
    start_credit_timer();
  }, 1000);
}, 1000);

