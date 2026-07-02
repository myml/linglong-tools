var token = localStorage.getItem('linglong_token') || '';
var repos = [];
var apps = [];
var currentPage = 1;
var pageSize = 20;

function api(method, path, body, headers) {
    var opts = {
        method: method,
        headers: Object.assign({ 'Content-Type': 'application/json' }, headers || {})
    };
    if (token) {
        opts.headers['X-Token'] = token;
    }
    if (body && method !== 'GET' && method !== 'HEAD') {
        if (body instanceof FormData) {
            delete opts.headers['Content-Type'];
            opts.body = body;
        } else {
            opts.body = JSON.stringify(body);
        }
    }
    return fetch(path, opts).then(function(r) { return r.json(); });
}

function showToast(msg, type) {
    type = type || 'info';
    var el = document.getElementById('toast');
    el.textContent = msg;
    el.className = 'toast ' + type + ' show';
    setTimeout(function() { el.className = 'toast ' + type; }, 3000);
}

function showLoginModal() {
    if (token) {
        if (confirm('已登录，是否退出登录？')) {
            token = '';
            localStorage.removeItem('linglong_token');
            updateLoginStatus();
            showToast('已退出登录', 'info');
        }
        return;
    }
    document.getElementById('loginModal').style.display = 'flex';
}

function closeModal(id) {
    document.getElementById(id).style.display = 'none';
}

function doLogin() {
    var username = document.getElementById('loginUsername').value.trim();
    var password = document.getElementById('loginPassword').value.trim();
    if (!username || !password) {
        showToast('请输入用户名和密码', 'error');
        return;
    }
    api('POST', '/api/v1/sign-in', { username: username, password: password })
        .then(function(res) {
            if (res.code === 0 || res.data) {
                token = res.data.token || '';
                localStorage.setItem('linglong_token', token);
                closeModal('loginModal');
                updateLoginStatus();
                showToast('登录成功', 'success');
            } else {
                showToast(res.msg || '登录失败', 'error');
            }
        })
        .catch(function(err) {
            showToast('登录请求失败: ' + err.message, 'error');
        });
}

function updateLoginStatus() {
    var btn = document.getElementById('loginBtn');
    var status = document.getElementById('loginStatus');
    if (token) {
        btn.textContent = '退出';
        status.textContent = '已登录';
    } else {
        btn.textContent = '登录';
        status.textContent = '';
    }
}

function loadRepos() {
    var container = document.getElementById('repoList');
    container.innerHTML = '<div class="loading">加载中</div>';

    api('GET', '/api/v1/repos')
        .then(function(res) {
            if (res.code !== 0 && res.code !== undefined && !res.data) {
                showToast(res.msg || '加载仓库失败', 'error');
                container.innerHTML = '<div class="empty-state">加载失败</div>';
                return;
            }
            repos = res.data || [];
            renderRepoList(repos);
            updateSearchRepoSelect(repos);
            updateUploadRepoSelect(repos);
        })
        .catch(function(err) {
            showToast('请求失败: ' + err.message, 'error');
            container.innerHTML = '<div class="empty-state">请求失败</div>';
        });
}

function renderRepoList(list) {
    var container = document.getElementById('repoList');
    if (!list || list.length === 0) {
        container.innerHTML = '<div class="empty-state">暂无仓库</div>';
        return;
    }
    var html = '';
    list.forEach(function(repo) {
        var refCount = repo.refs ? repo.refs.length : 0;
        html += '<div class="repo-item" data-repo="' + escapeHtml(repo.name) + '" onclick="selectRepo(\'' + escapeHtml(repo.name) + '\')">'
            + '<div class="repo-name">' + escapeHtml(repo.name) + '</div>'
            + '<div class="repo-mode">模式: ' + escapeHtml(repo.mode || '-') + '</div>'
            + '<div class="repo-refs-count">Refs: ' + refCount + '</div>'
            + '</div>';
    });
    container.innerHTML = html;
}

function selectRepo(repoName) {
    var items = document.querySelectorAll('.repo-item');
    items.forEach(function(el) { el.classList.remove('active'); });
    var active = document.querySelector('.repo-item[data-repo="' + repoName + '"]');
    if (active) active.classList.add('active');

    document.getElementById('searchRepo').value = repoName;

    searchApps();
}

function updateSearchRepoSelect(repoList) {
    var sel = document.getElementById('searchRepo');
    var current = sel.value;
    sel.innerHTML = '<option value="">选择仓库</option>';
    repoList.forEach(function(repo) {
        var opt = document.createElement('option');
        opt.value = repo.name;
        opt.textContent = repo.name;
        sel.appendChild(opt);
    });
    if (current) sel.value = current;
}

function updateUploadRepoSelect(repoList) {
    var sel = document.getElementById('uploadRepo');
    sel.innerHTML = '';
    repoList.forEach(function(repo) {
        var opt = document.createElement('option');
        opt.value = repo.name;
        opt.textContent = repo.name;
        sel.appendChild(opt);
    });
}

function searchApps() {
    var repoName = document.getElementById('searchRepo').value;
    var appId = document.getElementById('searchAppId').value.trim();
    var channel = document.getElementById('searchChannel').value;
    var arch = document.getElementById('searchArch').value;
    var version = document.getElementById('searchVersion').value.trim();
    var module = document.getElementById('searchModule').value.trim();

    if (!repoName) {
        showToast('请选择仓库', 'error');
        return;
    }

    var tableWrapper = document.getElementById('appTable');
    tableWrapper.innerHTML = '<div class="loading">搜索中</div>';

    var hasSearchField = appId || channel || arch || version || module;

    if (hasSearchField) {
        var params = new URLSearchParams();
        params.set('repo_name', repoName);
        if (channel) params.set('channel', channel);
        else params.set('channel', '');
        if (appId) params.set('app_id', appId);
        else params.set('app_id', '');
        if (arch) params.set('arch', arch);
        else params.set('arch', '');
        if (module) params.set('module', module);
        else params.set('module', '');
        if (version) params.set('version', version);

        api('GET', '/api/v2/search/apps?' + params.toString())
            .then(function(res) {
                apps = res.data || [];
                currentPage = 1;
                renderAppTable();
            })
            .catch(function(err) {
                showToast('搜索失败: ' + err.message, 'error');
                tableWrapper.innerHTML = '<div class="empty-state">搜索失败</div>';
            });
    } else {
        api('POST', '/api/v0/apps/fuzzysearchapp', { repoName: repoName })
            .then(function(res) {
                apps = res.data || [];
                currentPage = 1;
                renderAppTable();
            })
            .catch(function(err) {
                showToast('搜索失败: ' + err.message, 'error');
                tableWrapper.innerHTML = '<div class="empty-state">搜索失败</div>';
            });
    }
}

function renderAppTable() {
    var tableWrapper = document.getElementById('appTable');
    var pagination = document.getElementById('pagination');

    if (!apps || apps.length === 0) {
        tableWrapper.innerHTML = '<div class="empty-state">未找到应用</div>';
        pagination.style.display = 'none';
        return;
    }

    var totalPages = Math.ceil(apps.length / pageSize);
    var start = (currentPage - 1) * pageSize;
    var end = Math.min(start + pageSize, apps.length);
    var pageApps = apps.slice(start, end);

    var html = '<table>'
        + '<thead><tr>'
        + '<th>应用ID</th>'
        + '<th>名称</th>'
        + '<th>版本</th>'
        + '<th>频道</th>'
        + '<th>架构</th>'
        + '<th>模块</th>'
        + '<th>大小</th>'
        + '<th>操作</th>'
        + '</tr></thead>'
        + '<tbody>';

    pageApps.forEach(function(app, idx) {
        var sizeStr = app.size ? formatSize(app.size) : '-';
        html += '<tr>'
            + '<td>' + escapeHtml(app.appId || app.id || '-') + '</td>'
            + '<td>' + escapeHtml(app.name || '-') + '</td>'
            + '<td>' + escapeHtml(app.version || '-') + '</td>'
            + '<td><span class="tag tag-channel">' + escapeHtml(app.channel || '-') + '</span></td>'
            + '<td><span class="tag tag-arch">' + escapeHtml(app.arch || '-') + '</span></td>'
            + '<td><span class="tag tag-module">' + escapeHtml(app.module || '-') + '</span></td>'
            + '<td>' + sizeStr + '</td>'
            + '<td><button class="btn btn-sm btn-danger" onclick="confirmDelete(' + (start + idx) + ')">删除</button></td>'
            + '</tr>';
    });

    html += '</tbody></table>';
    tableWrapper.innerHTML = html;

    if (totalPages > 1) {
        var pagHtml = '<button onclick="goPage(' + (currentPage - 1) + ')" ' + (currentPage <= 1 ? 'disabled' : '') + '>上一页</button>';
        var startPage = Math.max(1, currentPage - 3);
        var endPage = Math.min(totalPages, startPage + 6);
        for (var i = startPage; i <= endPage; i++) {
            pagHtml += '<button onclick="goPage(' + i + ')" class="' + (i === currentPage ? 'active' : '') + '">' + i + '</button>';
        }
        pagHtml += '<button onclick="goPage(' + (currentPage + 1) + ')" ' + (currentPage >= totalPages ? 'disabled' : '') + '>下一页</button>';
        pagHtml += '<span style="font-size:12px;color:#999;margin-left:8px;">共 ' + apps.length + ' 条</span>';
        pagination.innerHTML = pagHtml;
        pagination.style.display = 'flex';
    } else {
        pagination.style.display = 'none';
    }
}

function goPage(p) {
    var totalPages = Math.ceil(apps.length / pageSize);
    if (p < 1 || p > totalPages) return;
    currentPage = p;
    renderAppTable();
}

function formatSize(bytes) {
    if (bytes === 0) return '0 B';
    var units = ['B', 'KB', 'MB', 'GB'];
    var i = Math.floor(Math.log(bytes) / Math.log(1024));
    return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i];
}

function confirmDelete(appIndex) {
    var app = apps[appIndex];
    if (!app) return;

    var text = '确定要删除以下应用吗？\n\n'
        + '应用ID: ' + (app.appId || app.id) + '\n'
        + '版本: ' + (app.version || '-') + '\n'
        + '频道: ' + (app.channel || '-') + '\n'
        + '架构: ' + (app.arch || '-') + '\n'
        + '模块: ' + (app.module || '-');

    document.getElementById('confirmText').textContent = text;
    document.getElementById('confirmHardDelete').checked = false;
    document.getElementById('confirmModal').style.display = 'flex';

    document.getElementById('confirmDeleteBtn').onclick = function() {
        doDelete(app);
    };
}

function doDelete(app) {
    if (!token) {
        showToast('请先登录', 'error');
        closeModal('confirmModal');
        showLoginModal();
        return;
    }

    var repoName = document.getElementById('searchRepo').value;
    var channel = app.channel;
    var appId = app.appId || app.id;
    var version = app.version;
    var arch = app.arch;
    var mod = app.module;
    var hard = document.getElementById('confirmHardDelete').checked;

    var path = '/api/v1/repos/' + encodeURIComponent(repoName)
        + '/refs/' + encodeURIComponent(channel)
        + '/' + encodeURIComponent(appId)
        + '/' + encodeURIComponent(version)
        + '/' + encodeURIComponent(arch)
        + '/' + encodeURIComponent(mod);

    if (hard) {
        path += '?hard=true';
    }

    api('DELETE', path)
        .then(function(res) {
            closeModal('confirmModal');
            if (res.code === 0 || res.code === undefined) {
                showToast('删除成功', 'success');
                searchApps();
            } else {
                showToast(res.msg || '删除失败', 'error');
            }
        })
        .catch(function(err) {
            closeModal('confirmModal');
            showToast('删除失败: ' + err.message, 'error');
        });
}

function showUploadModal() {
    if (!token) {
        showToast('请先登录', 'error');
        showLoginModal();
        return;
    }
    document.getElementById('uploadRef').value = '';
    document.getElementById('uploadFile').value = '';
    document.getElementById('uploadProgress').style.display = 'none';
    document.getElementById('uploadSubmitBtn').disabled = false;
    document.getElementById('uploadModal').style.display = 'flex';
}

function doUpload() {
    if (!token) {
        showToast('请先登录', 'error');
        closeModal('uploadModal');
        showLoginModal();
        return;
    }

    var repoName = document.getElementById('uploadRepo').value;
    var ref = document.getElementById('uploadRef').value.trim();
    var fileInput = document.getElementById('uploadFile');
    var file = fileInput.files[0];

    if (!repoName) {
        showToast('请选择目标仓库', 'error');
        return;
    }
    if (!ref) {
        showToast('请输入Ref', 'error');
        return;
    }
    if (!file) {
        showToast('请选择上传文件', 'error');
        return;
    }

    document.getElementById('uploadSubmitBtn').disabled = true;
    document.getElementById('uploadProgress').style.display = 'block';
    updateUploadProgress(10, '正在创建上传任务...');

    api('POST', '/api/v1/upload-tasks', { repoName: repoName, ref: ref })
        .then(function(res) {
            if (!res.data || !res.data.id) {
                showToast(res.msg || '创建上传任务失败', 'error');
                document.getElementById('uploadSubmitBtn').disabled = false;
                return null;
            }
            var taskId = res.data.id;
            updateUploadProgress(30, '任务已创建，正在上传文件...');

            var formData = new FormData();
            formData.append('file', file);

            var fileName = file.name.toLowerCase();
            var uploadPath;
            if (fileName.endsWith('.uab') || fileName.endsWith('.layer')) {
                uploadPath = '/api/v1/upload-tasks/' + taskId + '/layer';
            } else {
                uploadPath = '/api/v1/upload-tasks/' + taskId + '/tar';
            }

            var opts = {
                method: 'PUT',
                headers: { 'X-Token': token },
                body: formData
            };

            return fetch(uploadPath, opts).then(function(r) { return r.json(); })
                .then(function(uploadRes) {
                    if (uploadRes.code !== 0 && uploadRes.code !== undefined && !uploadRes.data) {
                        showToast(uploadRes.msg || '上传失败', 'error');
                        document.getElementById('uploadSubmitBtn').disabled = false;
                        return;
                    }
                    updateUploadProgress(80, '文件已上传，正在检查状态...');
                    pollUploadStatus(taskId);
                });
        })
        .catch(function(err) {
            showToast('上传失败: ' + err.message, 'error');
            document.getElementById('uploadSubmitBtn').disabled = false;
        });
}

function pollUploadStatus(taskId) {
    var maxPolls = 30;
    var count = 0;

    function poll() {
        if (count >= maxPolls) {
            updateUploadProgress(100, '状态查询超时，请手动确认');
            showToast('上传状态查询超时', 'warning');
            closeModal('uploadModal');
            document.getElementById('uploadSubmitBtn').disabled = false;
            searchApps();
            return;
        }
        count++;

        api('GET', '/api/v1/upload-tasks/' + taskId + '/status')
            .then(function(res) {
                var status = (res.data && res.data.status) ? res.data.status : '';
                if (status === 'success' || status === 'ok' || status === 'finished' || status === 'complete') {
                    updateUploadProgress(100, '上传完成');
                    showToast('上传成功', 'success');
                    setTimeout(function() {
                        closeModal('uploadModal');
                        document.getElementById('uploadSubmitBtn').disabled = false;
                        searchApps();
                    }, 1000);
                } else if (status === 'failed' || status === 'error') {
                    updateUploadProgress(100, '上传失败');
                    showToast('上传处理失败', 'error');
                    document.getElementById('uploadSubmitBtn').disabled = false;
                } else {
                    var pct = 80 + Math.min(count * 2, 18);
                    updateUploadProgress(pct, '处理中... 状态: ' + (status || '等待中'));
                    setTimeout(poll, 2000);
                }
            })
            .catch(function() {
                setTimeout(poll, 2000);
            });
    }

    poll();
}

function updateUploadProgress(pct, text) {
    document.getElementById('uploadProgressBar').style.width = pct + '%';
    document.getElementById('uploadStatus').textContent = text;
}

function escapeHtml(str) {
    if (!str) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

document.addEventListener('DOMContentLoaded', function() {
    updateLoginStatus();
    loadRepos();

    document.getElementById('searchAppId').addEventListener('keydown', function(e) {
        if (e.key === 'Enter') searchApps();
    });

    document.querySelectorAll('.modal').forEach(function(modal) {
        modal.addEventListener('click', function(e) {
            if (e.target === modal) {
                modal.style.display = 'none';
            }
        });
    });
});
