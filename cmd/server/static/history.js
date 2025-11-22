// Retry utility with exponential backoff
async function fetchWithRetry(url, options = {}, maxRetries = 3, baseDelay = 1000) {
    let lastError;
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 30000);
            const response = await fetch(url, {
                ...options,
                signal: controller.signal
            });
            clearTimeout(timeoutId);
            
            if (!response.ok) {
                const errorText = await response.text().catch(() => '');
                if (response.status >= 500 && attempt < maxRetries) {
                    throw new Error(`Server error (${response.status}): ${errorText || response.statusText}`);
                }
                throw new Error(errorText || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            return response;
        } catch (error) {
            lastError = error;
            if (attempt < maxRetries) {
                const isConnectionError = error.name === 'AbortError' || 
                                        error.message.includes('Failed to fetch') ||
                                        error.message.includes('NetworkError') ||
                                        error.message.includes('timeout');
                
                if (isConnectionError || (error.message && error.message.includes('500'))) {
                    const delay = baseDelay * Math.pow(2, attempt);
                    console.log(`Request failed (attempt ${attempt + 1}/${maxRetries + 1}), retrying in ${delay}ms...`, error.message);
                    await new Promise(resolve => setTimeout(resolve, delay));
                    continue;
                }
            }
            throw error;
        }
    }
    throw lastError;
}

// Check if Steam connection is ready and wait if needed
async function waitForSteamConnection(maxWait = 30000, checkInterval = 1000) {
    const startTime = Date.now();
    while (Date.now() - startTime < maxWait) {
        try {
            const res = await fetch('/api/steam/status');
            const data = await res.json();
            if (data.status === 3 || data.status === 4) {
                return true;
            }
        } catch (err) {
            console.error('Error checking Steam status:', err);
        }
        await new Promise(resolve => setTimeout(resolve, checkInterval));
    }
    return false;
}

// History Logic
document.addEventListener('DOMContentLoaded', () => {
    const historySteamIdInput = document.getElementById('history-steam-id');
    const historyLimitInput = document.getElementById('history-limit');
    const historyTurboOnlyCheckbox = document.getElementById('history-turbo-only');
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
    
    if (savedHistorySteamId) historySteamIdInput.value = savedHistorySteamId;
    if (savedHistoryLimit) historyLimitInput.value = savedHistoryLimit;
    if (savedHistoryTurboOnly === 'true') historyTurboOnlyCheckbox.checked = true;

    historySteamIdInput.addEventListener('input', () => {
        localStorage.setItem('historySteamId', historySteamIdInput.value);
    });
    
    historyLimitInput.addEventListener('input', () => {
        localStorage.setItem('historyLimit', historyLimitInput.value);
    });

    historyTurboOnlyCheckbox.addEventListener('change', () => {
        localStorage.setItem('historyTurboOnly', historyTurboOnlyCheckbox.checked ? 'true' : 'false');
    });

    fetchHistoryBtn.addEventListener('click', async () => {
        const steamId = historySteamIdInput.value.trim();
        const limit = historyLimitInput.value;
        const turboOnly = historyTurboOnlyCheckbox.checked;
        
        if (!steamId) {
            alert('Please enter a Steam ID');
            return;
        }

        // Check Steam connection status first
        try {
            const statusRes = await fetchWithRetry('/api/steam/status');
            const statusData = await statusRes.json();
            // StatusConnected = 3, StatusGCReady = 4
            if (statusData.status !== 3 && statusData.status !== 4) {
                historyResults.innerHTML = '<p class="loading">Waiting for Steam connection...</p>';
                const connected = await waitForSteamConnection(30000);
                if (!connected) {
                    alert('Steam connection required. Please connect to Steam first.');
                    historyResults.innerHTML = '<p style="color:red">Steam connection timeout. Please connect to Steam and try again.</p>';
                    return;
                }
            }
        } catch (err) {
            console.error('Error checking Steam status:', err);
            historyResults.innerHTML = '<p style="color:red">Error checking Steam connection status. Retrying...</p>';
            const connected = await waitForSteamConnection(15000);
            if (!connected) {
                alert('Error checking Steam connection status.');
                historyResults.innerHTML = '<p style="color:red">Could not verify Steam connection.</p>';
                return;
            }
        }
        
        historyResults.innerHTML = '<p class="loading">Fetching match history...</p>';
        
        const turboParam = turboOnly ? '&turboOnly=true' : '';
        fetchWithRetry(`/api/history?steamId=${steamId}&limit=${limit}${turboParam}`, {}, 3, 1000)
            .then(res => res.json())
            .then(matches => {
                renderHistory(matches);
                // Check existence after rendering, then auto-download
                checkReplayExistence(matches).then(existingIds => {
                    autoDownloadMatches(matches, existingIds);
                });
            })
            .catch(err => {
                historyResults.innerHTML = `<p style="color:red">Error: ${err.message}. <button onclick="location.reload()" style="margin-left: 10px; padding: 4px 8px;">Retry</button></p>`;
            });
    });

    function renderHistory(matches) {
        if (!matches || matches.length === 0) {
            historyResults.innerHTML = '<p>No matches found.</p>';
            return;
        }
        
        const profileName = getSelectedProfileName();
        const url = '/api/replays?t=' + Date.now() + (profileName ? '&profile=' + encodeURIComponent(profileName) : '');
        fetchWithRetry(url, {}, 2, 1000)
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
                        <span class="status-pill status-pending" id="status-${m.id}" style="visibility: hidden; min-width: 90px; text-align: center;">Checking...</span>
                        <button class="small-btn download-btn" data-match="${m.id}" style="min-width: 130px;">Download Replay</button>
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
                            <span class="status-pill status-pending" id="status-${m.id}" style="visibility: hidden; min-width: 90px; text-align: center;">Checking...</span>
                            <button class="small-btn download-btn" data-match="${m.id}" style="min-width: 130px;">Download Replay</button>
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
        fetchWithRetry(url, {}, 2, 1000)
            .then(res => res.json())
            .then(replays => {
                const safeReplays = replays || [];
                const existingIds = new Set(safeReplays.map(r => r.fileName.replace('.dem', '')));
                
                document.querySelectorAll('.history-item').forEach(item => {
                    const matchId = item.id.replace('history-match-', '');
                    const status = document.getElementById(`status-${matchId}`);
                    const btn = item.querySelector('.download-btn');
                    const exists = existingIds.has(matchId);
                    const currentlyExists = item.classList.contains('exists');
                    const statusVisible = status && status.style.visibility !== 'hidden';
                    
                    if (exists === currentlyExists && ((exists && statusVisible) || (!exists && !statusVisible))) {
                        return;
                    }
                    
                    if (exists) {
                        if (!currentlyExists) {
                            item.classList.add('exists');
                        }
                        if (btn && btn.textContent !== 'Redownload') { 
                            btn.textContent = 'Redownload'; 
                            btn.style.display = 'inline-block';
                            btn.disabled = false;
                        }
                        if (status) {
                            if (status.textContent !== 'Downloaded') {
                                status.textContent = 'Downloaded';
                            }
                            if (!status.classList.contains('status-success')) {
                                status.className = 'status-pill status-success';
                            }
                            if (status.style.visibility === 'hidden') {
                                status.style.visibility = 'visible';
                            }
                        }
                    } else {
                        if (currentlyExists) {
                            item.classList.remove('exists');
                        }
                        if (btn && btn.textContent !== 'Download Replay') { 
                            btn.textContent = 'Download Replay'; 
                            btn.style.display = 'inline-block';
                            btn.disabled = false;
                        }
                        if (status && status.style.visibility !== 'hidden') {
                            status.style.visibility = 'hidden';
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
        return fetchWithRetry(url, {}, 2, 1000)
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
                            status.style.visibility = 'visible';
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
                status.style.visibility = 'visible';
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
            fetchWithRetry('/api/download', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    matchId: parseInt(matchId),
                    profileName: profileName
                })
            }, 3, 2000)
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
                        status.style.visibility = 'visible';
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

    // Fatal search functionality
    const fatalSteamIdInput = document.getElementById('fatal-steam-id');
    const fatalMaxDepthInput = document.getElementById('fatal-max-depth');
    const findFatalGamesBtn = document.getElementById('find-fatal-games');
    const fatalResults = document.getElementById('fatal-results');

    // Auto-fill Steam ID from selected profile (handled in app.js profile select change handler)
    // Initial update on page load
    function updateFatalSteamId() {
        const profileSelect = document.getElementById('profile-select');
        if (!profileSelect || !fatalSteamIdInput) return;
        const index = profileSelect.value;
        if (index === '') {
            fatalSteamIdInput.value = '';
            return;
        }
        const savedProfiles = localStorage.getItem('steamProfiles');
        if (!savedProfiles) return;
        try {
            const profiles = JSON.parse(savedProfiles);
            if (profiles[index] && profiles[index].id) {
                fatalSteamIdInput.value = profiles[index].id;
            }
        } catch (e) {
            console.error('Failed to parse profiles:', e);
        }
    }
    
    // Initial update after a short delay to ensure app.js has loaded profiles
    setTimeout(updateFatalSteamId, 100);

    if (findFatalGamesBtn) {
        findFatalGamesBtn.addEventListener('click', async () => {
            const steamId = fatalSteamIdInput ? fatalSteamIdInput.value.trim() : '';
            const maxDepth = fatalMaxDepthInput ? parseInt(fatalMaxDepthInput.value) : 1;
            const fatalGamesPerInput = document.getElementById('fatal-games-per');
            const gamesPerFatal = fatalGamesPerInput ? parseInt(fatalGamesPerInput.value) : 2;
            const profileName = getSelectedProfileName();

            if (!steamId) {
                alert('Please enter a Steam ID');
                return;
            }

            if (maxDepth < 1 || maxDepth > 10) {
                alert('Max depth must be between 1 and 10');
                return;
            }

            if (gamesPerFatal < 1 || gamesPerFatal > 15) {
                alert('Games per fatal must be between 1 and 15');
                return;
            }

            // Check Steam connection status
            try {
                const statusRes = await fetchWithRetry('/api/steam/status');
                const statusData = await statusRes.json();
                // StatusConnected = 3, StatusGCReady = 4
                if (statusData.status !== 3 && statusData.status !== 4) {
                    if (fatalResults) {
                        fatalResults.innerHTML = '<p class="loading">Waiting for Steam connection...</p>';
                    }
                    const connected = await waitForSteamConnection(30000);
                    if (!connected) {
                        alert('Steam connection required. Please connect to Steam first.');
                        if (fatalResults) {
                            fatalResults.innerHTML = '<p style="color:red">Steam connection timeout. Please connect to Steam and try again.</p>';
                        }
                        return;
                    }
                }
            } catch (err) {
                console.error('Error checking Steam status:', err);
                if (fatalResults) {
                    fatalResults.innerHTML = '<p style="color:red">Error checking Steam connection status. Retrying...</p>';
                }
                const connected = await waitForSteamConnection(15000);
                if (!connected) {
                    alert('Error checking Steam connection status.');
                    if (fatalResults) {
                        fatalResults.innerHTML = '<p style="color:red">Could not verify Steam connection.</p>';
                    }
                    return;
                }
            }

            if (fatalResults) {
                fatalResults.innerHTML = '<p class="loading">Searching for fatal games...</p>';
            }

            fetchWithRetry('/api/fatal-search', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    steamId: steamId,
                    maxDepth: maxDepth,
                    gamesPerFatal: gamesPerFatal,
                    profileName: profileName
                })
            }, 3, 1000)
            .then(async res => {
                if (!res.ok) {
                    const text = await res.text();
                    throw new Error(text || 'Failed to search fatal games');
                }
                return res.json();
            })
            .then(data => {
                const matches = data.matches || [];
                if (fatalResults) {
                    if (matches.length === 0) {
                        fatalResults.innerHTML = '<p>No fatal games found.</p>';
                        return;
                    }

                    fatalResults.innerHTML = `
                        <p style="margin-bottom: 10px;"><strong>Found ${matches.length} fatal game(s):</strong></p>
                        <ul>
                            ${matches.map((match, index) => {
                                const matchId = match.fatalMatchId || match;
                                const singleDraftId = match.singleDraftMatchId || null;
                                const singleDraftDate = match.singleDraftDate || null;
                                const additionalMatchIds = match.additionalMatchIds || [];
                                const matchData = JSON.stringify({ matchId, singleDraftId, singleDraftDate, gamesPerFatal, steamId, additionalMatchIds });
                                return `
                                <li class="history-item" style="margin-bottom: 8px;" id="fatal-item-${matchId}">
                                    <div class="match-row">
                                        <span class="match-id">Fatal Match ${matchId}${singleDraftId ? ` (SD: ${singleDraftId})` : ''}</span>
                                    </div>
                                    <div class="match-actions" style="margin-top: 8px;">
                                        <button class="small-btn download-fatal-btn" data-match='${matchData}' id="download-btn-${matchId}">Download to Fatal</button>
                                        <div class="download-progress" id="progress-${matchId}" style="display: none; margin-top: 8px;">
                                            <div style="display: flex; align-items: center; gap: 8px;">
                                                <div class="spinner" style="width: 16px; height: 16px; border: 2px solid #334155; border-top-color: #22c55e; border-radius: 50%; animation: spin 0.8s linear infinite;"></div>
                                                <span class="progress-text" style="font-size: 0.85em; color: #94a3b8;">Downloading...</span>
                                            </div>
                                        </div>
                                    </div>
                                </li>
                            `;
                            }).join('')}
                        </ul>
                        <button class="btn secondary-btn" id="download-all-fatal" style="margin-top: 10px;">Download All to Fatal</button>
                    `;

                    // Add event listeners for individual downloads
                    document.querySelectorAll('.download-fatal-btn').forEach(btn => {
                        btn.addEventListener('click', (e) => {
                            const matchData = JSON.parse(e.target.getAttribute('data-match'));
                            const profile = getSelectedProfileName();
                            downloadFatalReplay(matchData, profile, e.target);
                        });
                    });

                    // Add event listener for download all
                    const downloadAllBtn = document.getElementById('download-all-fatal');
                    if (downloadAllBtn) {
                        downloadAllBtn.addEventListener('click', () => {
                            downloadAllFatalReplays(matches, profileName, gamesPerFatal);
                        });
                    }
                }
            })
            .catch(err => {
                if (fatalResults) {
                    fatalResults.innerHTML = `<p style="color:red">Error: ${err.message} <button onclick="document.getElementById('find-fatal-games').click()" style="margin-left: 10px; padding: 4px 8px;">Retry</button></p>`;
                }
            });
        });
    }

    function downloadFatalReplay(matchData, profileName, btn) {
        const matchId = matchData.matchId;
        const progressDiv = document.getElementById(`progress-${matchId}`);
        const progressText = progressDiv ? progressDiv.querySelector('.progress-text') : null;
        
        if (btn) {
            btn.disabled = true;
            btn.style.display = 'none';
        }
        if (progressDiv) {
            progressDiv.style.display = 'block';
        }
        if (progressText) {
            progressText.textContent = 'Downloading fatal match...';
        }

        const payload = {
            matchId: parseInt(matchData.matchId),
            profileName: profileName,
            fatal: true,
            gamesPerFatal: matchData.gamesPerFatal || 2
        };
        
        if (matchData.steamId) {
            payload.steamId = parseInt(matchData.steamId);
        }
        
        if (matchData.singleDraftId && matchData.singleDraftDate) {
            payload.singleDraftMatchId = parseInt(matchData.singleDraftId);
            payload.singleDraftDate = parseInt(matchData.singleDraftDate);
        }
        
        if (matchData.additionalMatchIds) {
            payload.additionalMatchIds = matchData.additionalMatchIds;
        }

        const startTime = Date.now();
        let progressInterval;
        const updateProgress = () => {
            if (progressText) {
                const elapsed = Math.floor((Date.now() - startTime) / 1000);
                if (elapsed < 5) {
                    progressText.textContent = 'Downloading fatal match...';
                } else if (elapsed < 15) {
                    progressText.textContent = 'Downloading singledraft match...';
                } else {
                    progressText.textContent = `Processing (${elapsed}s)...`;
                }
            }
        };
        progressInterval = setInterval(updateProgress, 1000);

        return fetchWithRetry('/api/download', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        }, 3, 2000)
        .then(async res => {
            if (progressInterval) clearInterval(progressInterval);
            let data;
            const contentType = res.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                data = await res.json();
            } else {
                const text = await res.text();
                try {
                    data = JSON.parse(text);
                } catch {
                    data = { success: false, error: text || `HTTP ${res.status}: ${res.statusText}` };
                }
            }
            if (!res.ok) {
                const errorMsg = data.error || data.message || `HTTP ${res.status}: ${res.statusText}`;
                throw new Error(errorMsg);
            }
            return data;
        })
        .then(data => {
            if (data.success) {
                if (progressText) {
                    progressText.textContent = '✓ Downloaded';
                    progressText.style.color = '#22c55e';
                }
                if (progressDiv) {
                    setTimeout(() => {
                        progressDiv.style.display = 'none';
                    }, 2000);
                }
                if (btn) {
                    btn.textContent = 'Downloaded';
                    btn.style.color = '#22c55e';
                    btn.style.display = 'inline-block';
                    btn.disabled = false;
                }
                if (window.browseDirectory) {
                    setTimeout(() => window.browseDirectory(''), 500);
                }
                return data;
            } else if (data.status === 'queued') {
                if (progressText) {
                    progressText.textContent = '⏳ Queued for parsing...';
                    progressText.style.color = '#fbbf24';
                }
                if (btn) {
                    btn.textContent = 'Queued';
                    btn.style.color = '#fbbf24';
                    btn.style.display = 'inline-block';
                    btn.disabled = false;
                }
                if (progressDiv) {
                    setTimeout(() => {
                        progressDiv.style.display = 'none';
                    }, 3000);
                }
                return data;
            } else {
                throw new Error(data.message || 'Unknown error');
            }
        })
        .catch(err => {
            clearInterval(progressInterval);
            if (progressText) {
                progressText.textContent = '✗ Failed: ' + err.message;
                progressText.style.color = '#ef4444';
            }
            if (btn) {
                btn.disabled = false;
                btn.textContent = 'Retry';
                btn.style.display = 'inline-block';
            }
            if (progressDiv) {
                setTimeout(() => {
                    progressDiv.style.display = 'none';
                }, 3000);
            }
            throw err;
        });
    }

    function downloadAllFatalReplays(matches, profileName, gamesPerFatal = 2) {
        const downloadAllBtn = document.getElementById('download-all-fatal');
        if (downloadAllBtn) {
            downloadAllBtn.disabled = true;
            downloadAllBtn.textContent = `Downloading ${matches.length}...`;
        }

        let completed = 0;
        matches.forEach((match, index) => {
            setTimeout(() => {
                const matchId = match.fatalMatchId || match;
                const singleDraftId = match.singleDraftMatchId || null;
                const singleDraftDate = match.singleDraftDate || null;
                const additionalMatchIds = match.additionalMatchIds || [];
                const fatalSteamIdInput = document.getElementById('fatal-steam-id');
                const steamId = fatalSteamIdInput ? fatalSteamIdInput.value.trim() : '';
                const matchData = { matchId, singleDraftId, singleDraftDate, gamesPerFatal, steamId, additionalMatchIds };
                
                downloadFatalReplay(matchData, profileName, null)
                .then(() => {
                    completed++;
                    if (downloadAllBtn) {
                        downloadAllBtn.textContent = `Downloading ${completed}/${matches.length}...`;
                    }
                    if (completed === matches.length) {
                        if (downloadAllBtn) {
                            const failedCount = matches.length - completed;
                            if (failedCount > 0) {
                                downloadAllBtn.textContent = `Completed (${failedCount} failed)`;
                                downloadAllBtn.style.color = '#fbbf24';
                            } else {
                                downloadAllBtn.textContent = 'All Downloaded';
                                downloadAllBtn.style.color = '#22c55e';
                            }
                        }
                        if (window.browseDirectory) {
                            setTimeout(() => window.browseDirectory(''), 500);
                        }
                    }
                })
                .catch(err => {
                    console.error(`Failed to download match ${matchId}:`, err);
                    completed++;
                    if (downloadAllBtn) {
                        downloadAllBtn.textContent = `Downloading ${completed}/${matches.length}...`;
                    }
                    if (completed === matches.length) {
                        if (downloadAllBtn) {
                            downloadAllBtn.textContent = 'Completed (some may have failed)';
                            downloadAllBtn.style.color = '#fbbf24';
                        }
                        if (window.browseDirectory) {
                            setTimeout(() => window.browseDirectory(''), 500);
                        }
                    }
                });
            }, index * 1000);
        });
    }
});
