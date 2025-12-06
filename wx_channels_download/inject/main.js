function __wx_channels_video_decrypt(t, e, p) {
  for (
    var r = new Uint8Array(t), n = 0;
    n < t.byteLength && e + n < p.decryptor_array.length;
    n++
  )
    r[n] ^= p.decryptor_array[n];
  return r;
}
window.VTS_WASM_URL =
  "https://res.wx.qq.com/t/wx_fed/cdn_libs/res/decrypt-video-core/1.3.0/wasm_video_decode.wasm";
window.MAX_HEAP_SIZE = 33554432;
var decryptor_array;
let decryptor;
/** t 是要解码的视频内容长度    e 是 decryptor_array 的长度 */
function wasm_isaac_generate(t, e) {
  decryptor_array = new Uint8Array(e);
  var r = new Uint8Array(Module.HEAPU8.buffer, t, e);
  decryptor_array.set(r.reverse());
  if (decryptor) {
    decryptor.delete();
  }
}
let loaded = false;
/** 获取 decrypt_array */
async function __wx_channels_decrypt(seed) {
  if (!loaded) {
    await __wx_load_script(
      "https://res.wx.qq.com/t/wx_fed/cdn_libs/res/decrypt-video-core/1.3.0/wasm_video_decode.js"
    );
    loaded = true;
  }
  await sleep();
  decryptor = new Module.WxIsaac64(seed);
  // 调用该方法时，会调用 wasm_isaac_generate 方法
  // 131072 是 decryptor_array 的长度
  decryptor.generate(131072);
  // decryptor.delete();
  // const r = Uint8ArrayToBase64(decryptor_array);
  // decryptor_array = undefined;
  return decryptor_array;
}
async function show_progress_or_loaded_size(response, downloadItemId, loadingInstance) {
  var content_length = response.headers.get("Content-Length");
  var chunks = [];
  var total_size = content_length ? parseInt(content_length, 10) : 0;
  if (total_size) {
    __wx_log({
      msg: `${total_size} Bytes`,
    });
  }
  var loaded_size = 0;
  var reader = response.body.getReader();
  var lastUpdateTime = 0;
  var updateInterval = 200; // 每200ms更新一次loading提示，避免闪烁
  
  while (true) {
    var { done, value } = await reader.read();
    if (done) {
      break;
    }
    chunks.push(value);
    loaded_size += value.length;
    
    var now = Date.now();
    var shouldUpdateLoading = (now - lastUpdateTime) >= updateInterval;
    
    if (total_size) {
      var progress = (loaded_size / total_size) * 100;
      
      // 更新下载列表中的进度（实时更新）
      if (downloadItemId && typeof window.update_download_item_progress === "function") {
        window.update_download_item_progress(downloadItemId, progress);
      }
      
      // 更新 loading 提示的百分比（节流更新，避免闪烁）
      if (shouldUpdateLoading && loadingInstance && typeof loadingInstance.update === "function") {
        var progressText = progress.toFixed(1) + "%";
        loadingInstance.update("下载中 " + progressText);
        lastUpdateTime = now;
      }
      
      __wx_log({
        replace: 1,
        msg: `${progress.toFixed(2)}%`,
      });
    } else {
      // 如果没有总大小，根据已下载大小估算进度（可选）
      if (downloadItemId && typeof window.update_download_item_progress === "function") {
        // 可以根据已下载大小显示一个估算进度，这里暂时不显示
      }
      
      // 显示已下载字节数（节流更新）
      if (shouldUpdateLoading && loadingInstance && typeof loadingInstance.update === "function") {
        var sizeText = (loaded_size / 1024 / 1024).toFixed(2) + " MB";
        loadingInstance.update("下载中 " + sizeText);
        lastUpdateTime = now;
      }
      
      __wx_log({
        replace: 1,
        msg: `${loaded_size} Bytes`,
      });
    }
  }
  // 下载完成，确保进度为100%
  if (downloadItemId && typeof window.update_download_item_progress === "function") {
    window.update_download_item_progress(downloadItemId, 100);
  }
  
  // 更新 loading 提示为完成
  if (loadingInstance && typeof loadingInstance.update === "function") {
    loadingInstance.update("下载完成");
  }
  
  var blob = new Blob(chunks);
  return blob;
}
/** 用于下载已经播放的视频内容 */
async function __wx_channels_download(profile, filename) {
  console.log("__wx_channels_download");
  const data = profile.data;
  const blob = new Blob(data, { type: "video/mp4" });
  await __wx_load_script("https://res.wx.qq.com/t/wx_fed/cdn_libs/res/FileSaver.min.js");
  saveAs(blob, filename + ".mp4");
}
/** 下载非加密视频 */
async function __wx_channels_download2(profile, filename) {
  console.log("__wx_channels_download2");
  const url = profile.url;
  await __wx_load_script("https://res.wx.qq.com/t/wx_fed/cdn_libs/res/FileSaver.min.js");
  const ins = __wx_channel_loading();
  try {
    // 尝试获取下载项ID（如果存在）
    var downloadItemId = profile.downloadItemId;
    const response = await fetch(url);
    const blob = await show_progress_or_loaded_size(response, downloadItemId, ins);
    __wx_log({
      ignore_prefix: 1,
      msg: "",
    });
    __wx_log({
      msg: "下载完成",
    });
    saveAs(blob, filename + ".mp4");
  } catch (err) {
    __wx_log({
      msg: "下载失败\n" + err.message,
    });
  }
  ins.hide();
}
/** 下载图片视频 */
async function __wx_channels_download3(profile, filename) {
  console.log("__wx_channels_download3");
  const files = profile.files;
  await __wx_load_script("https://res.wx.qq.com/t/wx_fed/cdn_libs/res/FileSaver.min.js");
  await __wx_load_script("https://res.wx.qq.com/t/wx_fed/cdn_libs/res/jszip.min.js");
  const zip = new JSZip();
  zip.file("contact.txt", JSON.stringify(profile.contact, null, 2));
  const folder = zip.folder("images");
  // console.log("files", files);
  const fetchPromises = files
    .map((f) => f.url)
    .map(async (url, index) => {
      const response = await fetch(url);
      const blob = await response.blob();
      folder.file(index + 1 + ".png", blob);
    });
  const ins = __wx_channel_loading();
  try {
    await Promise.all(fetchPromises);
    const content = await zip.generateAsync({ type: "blob" });
    saveAs(content, filename + ".zip");
  } catch (err) {
    __wx_log({
      msg: "下载失败\n" + err.message,
    });
  }
  ins.hide();
}
/** 下载加密视频 */
async function __wx_channels_download4(profile, opt) {
  var { filename, toMP3 } = opt;
  console.log("__wx_channels_download4");
  if (__wx_channels_config__.downloadLocalServerEnabled) {
    var fullname = filename + (toMP3 ? ".mp3" : ".mp4");
    var url = `http://${__wx_channels_config__.downloadLocalServerAddr}/download?url=${encodeURIComponent(profile.url)}&key=${profile.key}&filename=${encodeURIComponent(fullname)}&mp3=${Number(toMP3)}`;
    var a = document.createElement("a");
    a.href = url;
    a.download = fullname;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    return;
  }
  await __wx_load_script("https://res.wx.qq.com/t/wx_fed/cdn_libs/res/FileSaver.min.js");
  if (__wx_channels_config__.downloadPauseWhenDownload) {
    __wx_channels_pause_cur_video();
  }
  const ins = __wx_channel_loading();
  // 尝试获取下载项ID（如果存在）
  var downloadItemId = profile.downloadItemId;
  const response = await fetch(profile.url);
  const blob = await show_progress_or_loaded_size(response, downloadItemId, ins);
  __wx_log({
    ignore_prefix: 1,
    msg: "",
  });
  __wx_log({
    msg: "下载完成，开始解密",
  });
  var array = new Uint8Array(await blob.arrayBuffer());
  if (profile.key) {
    try {
      const r = await __wx_channels_decrypt(profile.key);
      // console.log("[]after __wx_channels_decrypt", r);
      profile.decryptor_array = r;
      if (profile.decryptor_array) {
        array = __wx_channels_video_decrypt(array, 0, profile);
      }
    } catch (err) {
      var tip = "前端解密失败，尝试使用 decrypt 命令解密";
      __wx_log({
        msg: tip,
      });
    }
  }
  const result = new Blob([array], { type: "video/mp4" });
  if (toMP3) {
    var audioCtx = new AudioContext();
    audioCtx.decodeAudioData(array.buffer, async function (buffer) {
      await __wx_load_script("https://res.wx.qq.com/t/wx_fed/cdn_libs/res/recorder.min.js");
      var blob = mediaBufferToWav(buffer);
      var [err, data] = await wavBlobToMP3(blob);
      if (err) {
        alert(err.message);
        return;
      }
      saveAs(data, filename + ".mp3");
    });
  } else {
    saveAs(result, filename + ".mp4");
  }
  ins.hide();
  if (__wx_channels_config__.downloadPauseWhenDownload) {
    __wx_channels_play_cur_video();
  }
}
/** 下载为mp3 */
async function __wx_channels_download_as_mp3(profile, filename) {
  console.log("__wx_channels_download_as_mp3");
  if (!__wx_channels_config__.downloadLocalServerEnabled) {
    alert("请先开启本地下载服务");
    return;
  }
  const url = `http://${__wx_channels_config__.downloadLocalServerAddr}/download?url=${encodeURIComponent(profile.url)}&key=${profile.key}&mp3=1&filename=${encodeURIComponent(filename + ".mp3")}`;
  window.open(url);
}
/** 复制当前页面地址 */
function __wx_channels_handle_copy__() {
  __wx_channels_copy(location.href);
  if (window.__wx_channels_tip__ && window.__wx_channels_tip__.toast) {
    window.__wx_channels_tip__.toast("复制成功", 1e3);
  }
}
async function __wx_channels_handle_log__() {
  await __wx_load_script(
    "https://res.wx.qq.com/t/wx_fed/cdn_libs/res/FileSaver.min.js"
  );
  const content = document.body.innerHTML;
  const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
  saveAs(blob, "log.txt");
}
/** 下载指定规格的视频 */
async function __wx_channels_handle_click_download__(spec, mp3) {
  var profile = __wx_channels_store__.profile;
  if (!profile) {
    alert("检测不到视频，请将本工具更新到最新版");
    return;
  }
  const _profile = { ...profile };
  // 保留 downloadItemId（如果存在）
  if (profile.downloadItemId) {
    _profile.downloadItemId = profile.downloadItemId;
  }
  var filename = __wx_build_filename(profile, spec, __wx_channels_config__.downloadFilenameTemplate);
  if (!filename) {
    alert("文件名生成失败");
    return;
  }
  if (spec) {
    _profile.url = profile.url + "&X-snsvideoflag=" + spec.fileFormat;
  }
  __wx_log({
    msg: `${filename}
${location.href}

${_profile.url}
${_profile.key || "该视频未加密"}`,
  });
  if (_profile.type === "picture") {
    __wx_channels_download3(_profile, filename);
    return;
  }
  _profile.data = __wx_channels_store__.buffers;
  __wx_channels_download4(_profile, { filename, toMP3: mp3 });
}
/** 下载已加载的视频 */
function __wx_channels_download_cur__() {
  var profile = __wx_channels_store__.profile;
  if (!profile) {
    alert("检测不到视频，请将本工具更新到最新版");
    return;
  }
  if (__wx_channels_store__.buffers.length === 0) {
    alert("没有可下载的内容");
    return;
  }
  var filename = __wx_build_filename(profile, null, __wx_channels_config__.downloadFilenameTemplate);
  if (!filename) {
    alert("文件名生成失败");
    return;
  }
  profile.data = __wx_channels_store__.buffers;
  __wx_channels_download(profile, filename);
}
/** 打印下载原始文件命令 */
function __wx_channels_handle_print_download_command() {
  var profile = __wx_channels_store__.profile;
  if (!profile) {
    alert("检测不到视频，请将本工具更新到最新版");
    return;
  }
  var _profile = { ...profile };
  var filename = __wx_build_filename(_profile, null, __wx_channels_config__.downloadFilenameTemplate);
  if (!filename) {
    alert("文件名生成失败");
    return;
  }
  var command = `download --url "${_profile.url}"`;
  if (_profile.key) {
    command += ` --key ${_profile.key}`;
  }
  command += ` --filename "${filename}.mp4"`;
  __wx_log({
    msg: command,
  });
  if (window.__wx_channels_tip__ && window.__wx_channels_tip__.toast) {
    window.__wx_channels_tip__.toast("请在终端查看下载命令", 1e3);
  }
}
/** 下载视频封面 */
async function __wx_channels_handle_download_cover() {
  var profile = __wx_channels_store__.profile;
  if (!profile) {
    alert("检测不到视频，请将本工具更新到最新版");
    return;
  }
  var filename = __wx_build_filename(profile, null, __wx_channels_config__.downloadFilenameTemplate);
  if (!filename) {
    alert("文件名生成失败");
    return;
  }
  await __wx_load_script("https://res.wx.qq.com/t/wx_fed/cdn_libs/res/FileSaver.min.js");
  __wx_log({
    msg: `下载封面\n${profile.coverUrl}`,
  });
  const ins = __wx_channel_loading();
  try {
    const url = profile.coverUrl.replace(/^http/, "https");
    const response = await fetch(url);
    const blob = await response.blob();
    saveAs(blob, filename + ".jpg");
  } catch (err) {
    alert(err.message);
  }
  ins.hide();
}
var __wx_channels_tip__ = {};
var __wx_channels_store__ = {
  profile: null,
  profiles: [],
  keys: {},
  buffers: [],
};

// 隐藏三个点按钮
function hide_three_dots_button() {
  var style = document.createElement("style");
  style.textContent = `
    .op-more-btn,
    .context-menu__wrp.item-gap-combine.op-more-btn,
    [class*="op-more"],
    [class*="more-btn"] {
      display: none !important;
    }
  `;
  document.head.appendChild(style);
}

// 全局悬浮下载按钮（兜底，不依赖操作栏 DOM）
function insert_floating_download_btn() {
  if (document.getElementById("__wx_channels_floating_download_btn__")) {
    return;
  }
  
  // 创建按钮容器
  var btnContainer = document.createElement("div");
  btnContainer.id = "__wx_channels_floating_download_btn__";
  btnContainer.style.cssText =
    "position: fixed; right: 24px; top: 100px; z-index: 999999; " +
    "display: flex; align-items: center; gap: 8px;";
  
  // 创建主按钮
  var btn = document.createElement("div");
  btn.style.cssText =
    "background: linear-gradient(135deg, #07c160 0%, #06ad56 100%); " +
    "color: #fff; padding: 12px 20px; border-radius: 25px; " +
    "font-size: 15px; font-weight: 600; cursor: pointer; " +
    "box-shadow: 0 4px 12px rgba(7, 193, 96, 0.4), 0 2px 4px rgba(0,0,0,.1); " +
    "display: flex; align-items: center; gap: 8px; " +
    "transition: all 0.3s ease; user-select: none; " +
    "white-space: nowrap;";
  
  // 添加下载图标
  var icon = document.createElement("span");
  icon.innerHTML = "⬇️";
  icon.style.cssText = "font-size: 18px; line-height: 1;";
  
  // 添加文字
  var text = document.createElement("span");
  text.textContent = "下载当前视频";
  
  btn.appendChild(icon);
  btn.appendChild(text);
  
  // 悬停效果
  btn.onmouseenter = function() {
    this.style.transform = "scale(1.05)";
    this.style.boxShadow = "0 6px 16px rgba(7, 193, 96, 0.5), 0 2px 6px rgba(0,0,0,.15)";
  };
  btn.onmouseleave = function() {
    this.style.transform = "scale(1)";
    this.style.boxShadow = "0 4px 12px rgba(7, 193, 96, 0.4), 0 2px 4px rgba(0,0,0,.1)";
  };
  
  // 点击效果
  btn.onmousedown = function() {
    this.style.transform = "scale(0.98)";
  };
  btn.onmouseup = function() {
    this.style.transform = "scale(1.05)";
  };
  
  // 点击事件
  btn.onclick = function () {
    var store = window.__wx_channels_store__;
    if (!store || !store.profile) {
      __wx_log({
        msg: "没有视频数据",
      });
      return;
    }
    var spec = __wx_channels_config__.defaultHighest
      ? null
      : store.profile.spec[0];
    __wx_channels_handle_click_download__(spec);
  };
  
  // 添加脉冲动画（吸引注意）
  var pulseStyle = document.createElement("style");
  pulseStyle.textContent = `
    @keyframes pulse {
      0%, 100% {
        box-shadow: 0 4px 12px rgba(7, 193, 96, 0.4), 0 2px 4px rgba(0,0,0,.1);
      }
      50% {
        box-shadow: 0 4px 12px rgba(7, 193, 96, 0.6), 0 2px 4px rgba(0,0,0,.1), 0 0 0 8px rgba(7, 193, 96, 0.1);
      }
    }
    #__wx_channels_floating_download_btn__ > div:first-child {
      animation: pulse 2s ease-in-out infinite;
    }
  `;
  document.head.appendChild(pulseStyle);
  
  btnContainer.appendChild(btn);
  document.body.appendChild(btnContainer);
  
  // 延迟移除动画（3秒后停止，避免干扰）
  setTimeout(function() {
    if (btn.style) {
      btn.style.animation = "none";
    }
  }, 3000);
}

var __wx_channels_video_download_btn__ = icon_download1();
__wx_channels_video_download_btn__.onclick = () => {
  if (!window.__wx_channels_store__.profile) {
    __wx_log({
      msg: "没有视频数据",
    });
    return;
  }
  var spec = __wx_channels_config__.defaultHighest ? null : window.__wx_channels_store__.profile.spec[0];
  __wx_channels_handle_click_download__(spec);
};

async function __insert_download_btn_to_home_page() {
  var $container = await __wx_find_elm(function () {
    return document.querySelector(".slides-scroll");
  });
  if (!$container) {
    return;
  }
  var cssText = $container.style.cssText;
  var re = /translate3d\([0-9]{1,}px, {0,1}-{0,1}([0-9]{1,})%/;
  var matched = cssText.match(re);
  var idx = matched ? Number(matched[1]) / 100 : 0;
  var $item = document.querySelectorAll(".slides-item")[idx];
  var $existing_download_btn = $item.querySelector(".download-icon");
  if ($existing_download_btn) {
    return;
  }
  var $elm3 = await __wx_find_elm(function () {
    return $item.getElementsByClassName("click-box op-item")[0];
  });
  if (!$elm3) {
    return;
  }
  const $parent = $elm3.parentElement;
  if ($parent) {
    __wx_channels_video_download_btn__ = icon_download2();
    __wx_channels_video_download_btn__.onclick = () => {
      if (!window.__wx_channels_store__.profile) {
        __wx_log({
          msg: "没有视频数据",
        });
        return;
      }
      var spec = __wx_channels_config__.defaultHighest ? null : window.__wx_channels_store__.profile.spec[0];
      __wx_channels_handle_click_download__(spec);
    };
    $parent.appendChild(__wx_channels_video_download_btn__);
    __wx_log({
      msg: "注入下载按钮成功!",
    });
    return true;
  }
}

async function insert_download_btn() {
  __wx_log({
    msg: "等待注入下载按钮",
  });
  if (window.location.pathname.includes("/pages/home")) {
    var success = await __insert_download_btn_to_home_page();
    if (success) {
      return;
    }
  }
  var $elm2 = await __wx_find_elm(function () {
    return document.getElementsByClassName("full-opr-wrp layout-col")[0];
  });
  if ($elm2) {
    __wx_channels_video_download_btn__ = icon_download1();
    __wx_channels_video_download_btn__.onclick = () => {
      if (!window.__wx_channels_store__.profile) {
        __wx_log({
          msg: "没有视频数据",
        });
        return;
      }
      var spec = __wx_channels_config__.defaultHighest ? null : window.__wx_channels_store__.profile.spec[0];
      __wx_channels_handle_click_download__(spec);
    };
    var relative_node = $elm2.children[$elm2.children.length - 1];
    if (!relative_node) {
      __wx_log({
        msg: "注入下载按钮成功3!",
      });
      $elm2.appendChild(__wx_channels_video_download_btn__);
      return;
    }
    __wx_log({
      msg: "注入下载按钮成功4!",
    });
    $elm2.insertBefore(__wx_channels_video_download_btn__, relative_node);
    return;
  }
  var $elm1 = await __wx_find_elm(function () {
    return document.getElementsByClassName("full-opr-wrp layout-row")[0];
  });
  if ($elm1) {
    var relative_node = $elm1.children[$elm1.children.length - 1];
    if (!relative_node) {
      __wx_log({
        msg: "注入下载按钮成功1!",
      });
      $elm1.appendChild(__wx_channels_video_download_btn__);
      return;
    }
    __wx_log({
      msg: "注入下载按钮成功2!",
    });
    $elm1.insertBefore(__wx_channels_video_download_btn__, relative_node);
    return;
  }
  __wx_log({
    msg: "没有找到操作栏，注入下载按钮失败\n",
  });
}
setTimeout(async () => {
  // insert_download_btn(); // 隐藏下载按钮
  insert_floating_download_btn();
  hide_three_dots_button(); // 隐藏三个点按钮
}, 800);
