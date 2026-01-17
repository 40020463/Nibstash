// Tidemark - front-end interactions
document.addEventListener('DOMContentLoaded', function() {
    // Auto-dismiss alerts
    const alerts = document.querySelectorAll('.alert');
    alerts.forEach(function(alert) {
        setTimeout(function() {
            alert.style.opacity = '0';
            setTimeout(function() {
                alert.remove();
            }, 300);
        }, 3000);
    });

    // Confirm deletes
    const deleteLinks = document.querySelectorAll('a[href*="/delete"]');
    deleteLinks.forEach(function(link) {
        if (!link.onclick) {
            link.onclick = function(e) {
                if (!confirm('确定要删除吗？')) {
                    e.preventDefault();
                    return false;
                }
            };
        }
    });

    // Clear all bookmarks confirmation
    const clearAllButton = document.getElementById('clear-all-bookmarks');
    if (clearAllButton) {
        clearAllButton.addEventListener('click', function(e) {
            if (!confirm('确定要清空全部收藏吗？此操作不可恢复。')) {
                e.preventDefault();
                return;
            }
            const confirmText = prompt('请输入“清空”以确认');
            if (confirmText !== '清空') {
                e.preventDefault();
                alert('已取消清空');
            }
        });
    }

    // Form validation
    const forms = document.querySelectorAll('form');
    forms.forEach(function(form) {
        form.addEventListener('submit', function(e) {
            const requiredInputs = form.querySelectorAll('[required]');
            let valid = true;
            requiredInputs.forEach(function(input) {
                if (!input.value.trim()) {
                    valid = false;
                    input.classList.add('error');
                } else {
                    input.classList.remove('error');
                }
            });
            if (!valid) {
                e.preventDefault();
                alert('请填写必填项');
            }
        });
    });

    // Import progress overlay
    const importForm = document.querySelector('.import-form');
    const importProgress = document.getElementById('import-progress');
    if (importForm && importProgress) {
        importForm.addEventListener('submit', function(e) {
            if (e.defaultPrevented || !importForm.checkValidity()) {
                return;
            }
            importProgress.classList.add('active');
            const submitButton = importForm.querySelector('button[type="submit"]');
            if (submitButton) {
                submitButton.disabled = true;
                submitButton.textContent = '导入中...';
            }
        });
    }

    // Bulk selection actions
    const selectPageButton = document.getElementById('select-page');
    const deleteSelectedButton = document.getElementById('delete-selected');
    const bulkForm = document.getElementById('bulk-actions-form');
    const selectionCheckboxes = Array.from(document.querySelectorAll('.bookmark-select-input'));
    const updateSelectionState = function() {
        const anyChecked = selectionCheckboxes.some(function(cb) {
            return cb.checked;
        });
        if (deleteSelectedButton) {
            deleteSelectedButton.disabled = !anyChecked;
        }
    };
    const updateSelectButtonLabel = function() {
        if (!selectPageButton) {
            return;
        }
        const allChecked = selectionCheckboxes.length > 0 && selectionCheckboxes.every(function(cb) {
            return cb.checked;
        });
        selectPageButton.textContent = allChecked ? '取消全选' : '全选本页';
    };
    if (selectPageButton) {
        selectPageButton.addEventListener('click', function() {
            const allChecked = selectionCheckboxes.length > 0 && selectionCheckboxes.every(function(cb) {
                return cb.checked;
            });
            selectionCheckboxes.forEach(function(cb) {
                cb.checked = !allChecked;
            });
            updateSelectButtonLabel();
            updateSelectionState();
        });
    }
    if (selectionCheckboxes.length > 0) {
        selectionCheckboxes.forEach(function(cb) {
            cb.addEventListener('change', function() {
                updateSelectButtonLabel();
                updateSelectionState();
            });
        });
        updateSelectionState();
        updateSelectButtonLabel();
    }
    if (bulkForm) {
        bulkForm.addEventListener('submit', function(e) {
            const anyChecked = selectionCheckboxes.some(function(cb) {
                return cb.checked;
            });
            if (!anyChecked) {
                e.preventDefault();
                return;
            }
            if (!confirm('确定要删除选中的收藏吗？')) {
                e.preventDefault();
            }
        });
    }

    const clearFolderForm = document.getElementById('clear-folder-form');
    if (clearFolderForm) {
        clearFolderForm.addEventListener('submit', function(e) {
            if (!confirm('确定要删除当前文件夹下的所有收藏吗？此操作不可恢复。')) {
                e.preventDefault();
            }
        });
    }
});

// Copy to clipboard
function copyToClipboard(text) {
    if (navigator.clipboard) {
        navigator.clipboard.writeText(text).then(function() {
            alert('已复制到剪贴板');
        }).catch(function() {
            fallbackCopy(text);
        });
    } else {
        fallbackCopy(text);
    }
}

function fallbackCopy(text) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
    alert('已复制到剪贴板');
}
