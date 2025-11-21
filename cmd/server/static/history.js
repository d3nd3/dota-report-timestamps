// History Logic
document.addEventListener('DOMContentLoaded', () => {
    const historySteamIdInput = document.getElementById('history-steam-id');
    const historyLimitInput = document.getElementById('history-limit');
    const historyTurboOnlyCheckbox = document.getElementById('history-turbo-only');
    const historyUseGCCheckbox = document.getElementById('history-use-gc');
    const fetchHistoryBtn = document.getElementById('fetch-history');
    const historyResults = document.getElementById('history-results');

    function getSelectedProfileName() {
        const profileSelect = document.getElementById('profile-select');
        if (!profileSelect) return '';
        const index = profileSelect.value;
        if (index === '') return '';
        const savedProfiles = localStorage.getItem('steamProfiles');
        if (!savedProfiles) return '';
        try {
            const profiles = JSON.parse(savedProfiles);
            if (profiles[index]) {
                return profiles[index].name;
            }
        } catch (e) {
            console.error('Failed to parse profiles:', e);
        }
        return '';
    }

    window.getSelectedProfileName = getSelectedProfileName;

    // Load saved Steam ID for history
    const savedHistorySteamId = localStorage.getItem('historySteamId');
    const savedHistoryLimit = localStorage.getItem('historyLimit');
    const savedHistoryTurboOnly = localStorage.getItem('historyTurboOnly');
    const savedHistoryUseGC = localStorage.getItem('historyUseGC');
    
    if (savedHistorySteamId) historySteamIdInput.value = savedHistorySteamId;
    if (savedHistoryLimit) historyLimitInput.value = savedHistoryLimit;
    if (savedHistoryTurboOnly === 'true') historyTurboOnlyCheckbox.checked = true;
    if (savedHistoryUseGC === 'true') historyUseGCCheckbox.checked = true;

    historySteamIdInput.addEventListener('input', () => {
        localStorage.setItem('historySteamId', historySteamIdInput.value);
    });
    
    historyLimitInput.addEventListener('input', () => {
        localStorage.setItem('historyLimit', historyLimitInput.value);
    });

    historyTurboOnlyCheckbox.addEventListener('change', () => {
        localStorage.setItem('historyTurboOnly', historyTurboOnlyCheckbox.checked ? 'true' : 'false');
    });

    historyUseGCCheckbox.addEventListener('change', () => {
        localStorage.setItem('historyUseGC', historyUseGCCheckbox.checked ? 'true' : 'false');
    });

    fetchHistoryBtn.addEventListener('click', () => {
        const steamId = historySteamIdInput.value.trim();
        const limit = historyLimitInput.value;
        const turboOnly = historyTurboOnlyCheckbox.checked;
        const useGC = historyUseGCCheckbox.checked;
        
        if (!steamId) {
            alert('Please enter a Steam ID');
            return;
        }
        
        historyResults.innerHTML = '<p class="loading">Fetching match history...</p>';
        
        const turboParam = turboOnly ? '&turboOnly=true' : '';
        const gcParam = useGC ? '&useGC=true' : '';
        fetch(`/api/history?steamId=${steamId}&limit=${limit}${turboParam}${gcParam}`)
            .then(res => {
                if (!res.ok) throw new Error('Failed to fetch history');
                return res.json();
            })
            .then(matches => {
                renderHistory(matches);
                // Check existence after rendering, then auto-download
                checkReplayExistence(matches).then(existingIds => {
                    autoDownloadMatches(matches, existingIds);
                });
            })
            .catch(err => {
                historyResults.innerHTML = `<p style="color:red">Error: ${err.message}. Check API Token.</p>`;
            });
    });

    function renderHistory(matches) {
        if (!matches || matches.length === 0) {
            historyResults.innerHTML = '<p>No matches found.</p>';
            return;
        }
        
        const profileName = getSelectedProfileName();
        const url = '/api/replays?t=' + Date.now() + (profileName ? '&profile=' + encodeURIComponent(profileName) : '');
        fetch(url)
            .then(res => res.json())
            .then(replays => {
                const safeReplays = replays || [];
                const existingIds = new Set(safeReplays.map(r => r.fileName.replace('.dem', '')));
                
                historyResults.innerHTML = '<ul>' + matches.map(m => {
                    return `
                        <li class="history-item" id="history-match-${m.id}">
                            <div class="match-row">
                                <span class="match-id">Match ${m.id}</span>
                            </div>
                    <div class="match-actions" style="margin-top: 8px; display: flex; justify-content: space-between; align-items: center;">
                        <span class="status-pill status-pending" id="status-${m.id}" style="display: none;">Checking...</span>
                        <button class="small-btn download-btn" data-match="${m.id}">Download Replay</button>
                        <div class="progress-container hidden" id="progress-container-${m.id}" style="flex: 1; margin: 0 10px; height: 6px; background: #334155; border-radius: 3px;">
                             <div class="progress-bar" id="progress-${m.id}" style="width: 0%; height: 100%; background: #22c55e; border-radius: 3px; transition: width 0.2s;"></div>
                        </div>
                        </div>
                    </li>
                `;
            }).join('') + '</ul>';
            
            // Add event listeners
            document.querySelectorAll('.download-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const matchId = e.target.getAttribute('data-match');
                    downloadReplay(matchId, e.target);
                });
            });
        })
        .catch(err => {
            console.error('Error fetching replays:', err);
            historyResults.innerHTML = '<ul>' + matches.map(m => {
                return `
                    <li class="history-item" id="history-match-${m.id}">
                        <div class="match-row">
                            <span class="match-id">Match ${m.id}</span>
                        </div>
                        <div class="match-actions" style="margin-top: 8px; display: flex; justify-content: space-between; align-items: center;">
                            <span class="status-pill status-pending" id="status-${m.id}" style="display: none;">Checking...</span>
                            <button class="small-btn download-btn" data-match="${m.id}">Download Replay</button>
                            <div class="progress-container hidden" id="progress-container-${m.id}" style="flex: 1; margin: 0 10px; height: 6px; background: #334155; border-radius: 3px;">
                                 <div class="progress-bar" id="progress-${m.id}" style="width: 0%; height: 100%; background: #22c55e; border-radius: 3px; transition: width 0.2s;"></div>
                            </div>
                        </div>
                    </li>
                `;
            }).join('') + '</ul>';
            
            document.querySelectorAll('.download-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const matchId = e.target.getAttribute('data-match');
                    downloadReplay(matchId, e.target);
                });
            });
        });
    }
    
    window.updateHistoryStatus = function() {
        const profileName = getSelectedProfileName();
        const url = '/api/replays?t=' + Date.now() + (profileName ? '&profile=' + encodeURIComponent(profileName) : '');
        fetch(url)
            .then(res => res.json())
            .then(replays => {
                const safeReplays = replays || [];
                const existingIds = new Set(safeReplays.map(r => r.fileName.replace('.dem', '')));
                console.log('Updated history status. Found replays:', existingIds.size);
                
                document.querySelectorAll('.history-item').forEach(item => {
                    const matchId = item.id.replace('history-match-', '');
                    const status = document.getElementById(`status-${matchId}`);
                    const btn = item.querySelector('.download-btn');
                    
                    if (existingIds.has(matchId)) {
                        item.classList.add('exists');
                        if (btn) { 
                            btn.textContent = 'Redownload'; 
                            btn.style.display = 'inline-block';
                            btn.disabled = false;
                        }
                        if (status) {
                            status.textContent = 'Downloaded';
                            status.className = 'status-pill status-success';
                            status.style.display = 'inline-flex';
                        }
                    } else {
                        item.classList.remove('exists');
                        if (btn) { 
                            btn.textContent = 'Download Replay'; 
                            btn.style.display = 'inline-block';
                            btn.disabled = false;
                        }
                        if (status) {
                            status.style.display = 'none';
                            status.className = 'status-pill status-pending';
                        }
                    }
                });
            })
            .catch(err => console.error('Error updating history status:', err));
    };

    function checkReplayExistence(matches) {
        const profileName = getSelectedProfileName();
        const url = '/api/replays?t=' + Date.now() + (profileName ? '&profile=' + encodeURIComponent(profileName) : '');
        return fetch(url)
            .then(res => res.json())
            .then(replays => {
                const safeReplays = replays || [];
                const existingIds = new Set(safeReplays.map(r => r.fileName.replace('.dem', '')));
                matches.forEach(m => {
                    const btn = document.querySelector(`button[data-match="${m.id}"]`);
                    const status = document.getElementById(`status-${m.id}`);
                    
                    if (existingIds.has(m.id.toString())) {
                        const item = document.getElementById(`history-match-${m.id}`);
                        if (item) item.classList.add('exists');
                        if (btn) btn.textContent = 'Redownload';
                        if (status) {
                            status.textContent = 'Downloaded';
                            status.className = 'status-pill status-success';
                            status.style.display = 'inline-flex';
                        }
                    }
                });
                return existingIds;
            });
    }

    async function autoDownloadMatches(matches, existingIds) {
        const matchesToDownload = matches.filter(m => !existingIds.has(m.id.toString()));
        
        if (matchesToDownload.length === 0) {
            return;
        }

        const downloadStatus = document.getElementById('download-status');
        if (downloadStatus) {
            downloadStatus.classList.remove('hidden');
            downloadStatus.textContent = `Downloading ${matchesToDownload.length} match(es)...`;
        }

        for (let i = 0; i < matchesToDownload.length; i++) {
            const match = matchesToDownload[i];
            const btn = document.querySelector(`button[data-match="${match.id}"]`);
            
            if (downloadStatus) {
                downloadStatus.textContent = `Downloading ${i + 1} of ${matchesToDownload.length}: Match ${match.id}`;
            }

            try {
                await downloadReplayPromise(match.id, btn);
            } catch (err) {
                console.error(`Failed to download match ${match.id}:`, err);
            }
        }

        if (downloadStatus) {
            downloadStatus.textContent = `Completed downloading ${matchesToDownload.length} match(es)`;
            setTimeout(() => {
                downloadStatus.classList.add('hidden');
            }, 3000);
        }
    }

    function downloadReplayPromise(matchId, btn) {
        return new Promise((resolve, reject) => {
            if (btn && btn.disabled) {
                resolve();
                return;
            }
            
            const originalText = btn ? btn.textContent : '';
            if (btn) {
                btn.disabled = true;
                btn.textContent = 'Starting...';
            }
            
            const status = document.getElementById(`status-${matchId}`);
            if (status) {
                status.textContent = 'Downloading...';
                status.className = 'status-pill status-downloading';
                status.style.display = 'inline-flex';
            }

            const progressContainer = document.getElementById(`progress-container-${matchId}`);
            const progressBar = document.getElementById(`progress-${matchId}`);
            
            if (progressContainer && btn) {
                progressContainer.classList.remove('hidden');
                btn.style.display = 'none';
            }

            const eventSource = new EventSource(`/api/progress?matchId=${matchId}`);
            eventSource.onmessage = (event) => {
                const percent = parseFloat(event.data);
                if (progressBar) {
                    const currentWidth = parseFloat(progressBar.style.width) || 0;
                    if (percent >= currentWidth) {
                        progressBar.style.width = `${percent}%`;
                    }
                }
            };
            
            eventSource.onerror = () => {
                eventSource.close();
            };

            const profileName = getSelectedProfileName ? getSelectedProfileName() : '';
            fetch('/api/download', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    matchId: parseInt(matchId),
                    profileName: profileName
                })
            })
            .then(res => res.json())
            .then(data => {
                eventSource.close();
                if (progressBar) progressBar.style.width = '100%';
                
                return new Promise(resolve => setTimeout(() => resolve(data), 500));
            })
            .then(data => {
                if (data.success) {
                    const item = document.getElementById(`history-match-${matchId}`);
                    if (item) {
                        item.classList.add('exists');
                        item.classList.add('added-animation');
                    }
                    
                    if (status) {
                        status.textContent = 'Downloaded';
                        status.className = 'status-pill status-success';
                        status.style.display = 'inline-flex';
                    }
                    
                    if (window.loadReplays) {
                        setTimeout(() => window.loadReplays(), 500);
                    }
                    if (window.updateHistoryStatus) {
                        setTimeout(() => window.updateHistoryStatus(), 1000);
                    }
                    resolve(data);
                } else if (data.status === 'queued') {
                    if (status) {
                        status.textContent = 'Queued';
                        status.className = 'status-pill status-pending';
                    }
                    resolve(data);
                } else {
                    if (status) {
                        status.textContent = 'Failed';
                        status.className = 'status-pill status-error';
                    }
                    reject(new Error(data.message || 'Unknown error'));
                }
            })
            .catch(err => {
                eventSource.close();
                if (status) {
                    status.textContent = 'Error';
                    status.className = 'status-pill status-error';
                }
                reject(err);
            })
            .finally(() => {
                if (btn) {
                    btn.disabled = false;
                    btn.textContent = 'Redownload';
                    btn.style.display = 'inline-block';
                }
                if (progressContainer) progressContainer.classList.add('hidden');
                if (progressBar) {
                    progressBar.style.transition = 'none';
                    progressBar.style.width = '0%';
                    setTimeout(() => {
                        if (progressBar) progressBar.style.transition = 'width 0.2s';
                    }, 50);
                }
            });
        });
    }

    function downloadReplay(matchId, btn) {
        downloadReplayPromise(matchId, btn)
            .then(data => {
                if (data && data.status === 'queued') {
                    alert(`Match ${matchId} queued for parsing. It will be downloaded automatically in the background.`);
                } else if (data && !data.success) {
                    alert('Download failed: ' + (data.message || 'Unknown error'));
                }
            })
            .catch(err => {
                console.error(err);
                alert('Error downloading: ' + err.message);
            });
    }
});
