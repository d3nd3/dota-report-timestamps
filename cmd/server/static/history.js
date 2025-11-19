function downloadReplay(matchId, button) {
    const originalText = button.textContent;
    button.disabled = true;
    button.textContent = 'Requesting...';
    button.classList.add('loading-btn');

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10 * 60 * 1000); // 10 minute timeout

    fetch('/api/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ matchId: matchId }),
        signal: controller.signal
    })
    .then(async res => {
        clearTimeout(timeoutId);
        if (!res.ok) {
            const errorText = await res.text();
            throw new Error(errorText || res.statusText);
        }
        return res.json();
    })
    .then(data => {
        console.log('Download successful:', data);
        button.textContent = '✓ Downloaded';
        button.classList.remove('loading-btn');
        button.classList.add('success-btn');
        addMatchToReplayList(matchId);
        
        const replaceButtonWithStatus = () => {
            const matchIdStr = String(matchId).trim();
            const replayList = document.getElementById('replay-list');
            if (!replayList) return false;
            
            const checkboxes = replayList.querySelectorAll('input[type="checkbox"]');
            for (const cb of checkboxes) {
                if (String(cb.value).trim() === matchIdStr) {
                    const container = button.parentElement;
                    if (container && container.className === 'match-row') {
                        const statusSpan = document.createElement('span');
                        statusSpan.textContent = '✓ Added';
                        statusSpan.className = 'status-added';
                        statusSpan.style.color = 'var(--success-color)';
                        statusSpan.style.fontWeight = '600';
                        container.replaceChild(statusSpan, button);
                        return true;
                    }
                }
            }
            return false;
        };
        
        if (typeof window.loadReplays === 'function') {
            setTimeout(() => {
                window.loadReplays();
                setTimeout(() => {
                    if (!replaceButtonWithStatus()) {
                        button.textContent = originalText;
                        button.disabled = false;
                    }
                }, 1000);
            }, 500);
        } else if (typeof loadReplays === 'function') {
            setTimeout(() => {
                loadReplays();
                setTimeout(() => {
                    if (!replaceButtonWithStatus()) {
                        button.textContent = originalText;
                        button.disabled = false;
                    }
                }, 1000);
            }, 500);
        } else {
            const refreshBtn = document.getElementById('refresh-replays');
            if (refreshBtn) {
                setTimeout(() => {
                    refreshBtn.click();
                    setTimeout(() => {
                        if (!replaceButtonWithStatus()) {
                            button.textContent = originalText;
                            button.disabled = false;
                        }
                    }, 1000);
                }, 500);
            } else {
                setTimeout(() => {
                    if (!replaceButtonWithStatus()) {
                        button.textContent = originalText;
                        button.disabled = false;
                    }
                }, 2000);
            }
        }
    })
    .catch(err => {
        clearTimeout(timeoutId);
        button.textContent = 'Error';
        button.classList.remove('loading-btn');
        button.classList.add('danger-btn');
        let errorMsg = err.message;
        if (err.name === 'AbortError') {
            errorMsg = 'Request timed out. The download may still be processing. Check server logs.';
        }
        alert('Error downloading replay: ' + errorMsg);
        setTimeout(() => {
            button.textContent = originalText;
            button.classList.remove('danger-btn');
            button.classList.add('success-btn');
            button.disabled = false;
        }, 3000);
    });
}

function addMatchToReplayList(matchId) {
    const replayList = document.getElementById('replay-list');
    if (!replayList) return;

    const matchIdStr = String(matchId);
    const existingCheckbox = document.getElementById(`replay-${matchIdStr}`);
    if (existingCheckbox) {
        existingCheckbox.checked = true;
        return;
    }

    const div = document.createElement('div');
    div.className = 'replay-item';
    div.innerHTML = `
        <input type="checkbox" value="${matchIdStr}" id="replay-${matchIdStr}" checked>
        <label for="replay-${matchIdStr}">${matchIdStr}.dem <span style="color: var(--warning-color); font-size: 0.9em;">(downloaded)</span></label>
    `;
    // Add to top of list
    if (replayList.firstChild) {
        replayList.insertBefore(div, replayList.firstChild);
    } else {
        replayList.appendChild(div);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const historySteamIdInput = document.getElementById('history-steam-id');
    const historyLimitInput = document.getElementById('history-limit');
    const fetchHistoryBtn = document.getElementById('fetch-history');
    const downloadAllBtn = document.getElementById('download-all');
    const historyResults = document.getElementById('history-results');
    let currentMatches = [];

    if (!fetchHistoryBtn) return;

    // Load from localStorage
    const savedHistorySteamId = localStorage.getItem('historySteamId');
    const savedHistoryLimit = localStorage.getItem('historyLimit');
    if (savedHistorySteamId) historySteamIdInput.value = savedHistorySteamId;
    if (savedHistoryLimit) historyLimitInput.value = savedHistoryLimit;

    // Save to localStorage when values change
    historySteamIdInput.addEventListener('input', () => {
        if (historySteamIdInput.value) localStorage.setItem('historySteamId', historySteamIdInput.value);
    });
    historyLimitInput.addEventListener('input', () => {
        if (historyLimitInput.value) localStorage.setItem('historyLimit', historyLimitInput.value);
    });

    fetchHistoryBtn.addEventListener('click', () => {
        const steamId = historySteamIdInput.value;
        const limit = historyLimitInput.value;

        if (!steamId) {
            alert('Please enter a Steam ID');
            return;
        }

        // Save to localStorage
        localStorage.setItem('historySteamId', steamId);
        if (limit) localStorage.setItem('historyLimit', limit);

        historyResults.innerHTML = '<p class="loading">Fetching matches...</p>';

        function getExistingMatchIds() {
            const replayList = document.getElementById('replay-list');
            if (!replayList) {
                console.log('Replay list element not found');
                return new Set();
            }
            const checkboxes = replayList.querySelectorAll('input[type="checkbox"]');
            const existingIds = new Set();
            checkboxes.forEach(cb => {
                const matchId = String(cb.value).trim();
                if (matchId) {
                    existingIds.add(matchId);
                }
            });
            console.log('Found existing match IDs:', Array.from(existingIds));
            return existingIds;
        }

        fetch(`/api/history?steamId=${steamId}&limit=${limit}`)
            .then(async res => {
                if (!res.ok) {
                    const errorText = await res.text();
                    throw new Error(errorText || res.statusText);
                }
                return res.json();
            })
            .then(matches => {
                historyResults.innerHTML = '';
                currentMatches = matches || [];
                
                if (!matches || matches.length === 0) {
                    historyResults.innerHTML = '<p>No matches found.</p>';
                    downloadAllBtn.style.display = 'none';
                    return;
                }

                downloadAllBtn.style.display = 'inline-block';
                const existingMatchIds = getExistingMatchIds();

                const info = document.createElement('p');
                info.innerHTML = `Found ${matches.length} match(es). <strong>Note:</strong> Replay files (.dem) must already exist in your replay directory. Click match IDs to add them to the replay list.`;
                info.style.marginBottom = '10px';
                historyResults.appendChild(info);

                const ul = document.createElement('ul');
                matches.forEach(match => {
                    const matchIdStr = String(match.id).trim();
                    const alreadyExists = existingMatchIds.has(matchIdStr);
                    const li = document.createElement('li');
                    li.className = 'history-item';
                    if (alreadyExists) li.classList.add('exists');

                    const container = document.createElement('div');
                    container.className = 'match-row';
                    
                    const textNode = document.createElement('span');
                    textNode.className = 'match-id';
                    textNode.textContent = `Match ID: ${match.id}`;
                    container.appendChild(textNode);
                    
                    if (!alreadyExists) {
                        const downloadBtn = document.createElement('button');
                        downloadBtn.textContent = 'Download';
                        downloadBtn.className = 'btn success-btn small-btn';
                        downloadBtn.addEventListener('click', (e) => {
                            e.stopPropagation();
                            downloadReplay(match.id, downloadBtn);
                        });
                        container.appendChild(downloadBtn);
                    } else {
                        const statusSpan = document.createElement('span');
                        statusSpan.textContent = '✓ Added';
                        statusSpan.className = 'status-added';
                        statusSpan.style.color = 'var(--success-color)';
                        statusSpan.style.fontWeight = '600';
                        container.appendChild(statusSpan);
                    }
                    
                    li.appendChild(container);

                    li.addEventListener('click', () => {
                        addMatchToReplayList(match.id);
                        li.classList.add('added-animation'); // You can define this animation in CSS if you want
                        textNode.textContent = `Match ID: ${match.id} ✓ Added`;
                        
                        // Check if status span exists, if not add it temporarily or update text
                        // For simplicity, just visual feedback
                        
                        setTimeout(() => {
                            li.classList.remove('added-animation');
                            textNode.textContent = `Match ID: ${match.id}`;
                        }, 2000);
                    });
                    ul.appendChild(li);
                });
                historyResults.appendChild(ul);
            })
            .catch(err => {
                historyResults.innerHTML = `<p style="color:red">Error fetching history: ${err}</p>`;
                downloadAllBtn.style.display = 'none';
                currentMatches = [];
            });
    });

    function downloadAllMatches() {
        if (!currentMatches || currentMatches.length === 0) {
            alert('No matches to download. Please fetch matches first.');
            return;
        }

        const existingMatchIds = getExistingMatchIds();
        const matchesToDownload = currentMatches.filter(match => {
            const matchIdStr = String(match.id).trim();
            return !existingMatchIds.has(matchIdStr);
        });

        if (matchesToDownload.length === 0) {
            alert('All matches are already downloaded!');
            return;
        }

        const fileNames = matchesToDownload.map(m => m.id + '.dem').join(', ');
        if (!confirm(`Download ${matchesToDownload.length} replay file(s)?\n\n${fileNames}\n\nThis may take a while.`)) {
            return;
        }

        const originalText = downloadAllBtn.textContent;
        downloadAllBtn.disabled = true;
        downloadAllBtn.textContent = `Downloading... (0/${matchesToDownload.length})`;

        let completed = 0;
        let failed = 0;
        const errors = [];

        const downloadPromises = matchesToDownload.map((match, index) => {
            return new Promise((resolve) => {
                setTimeout(() => {
                    const matchId = match.id;
                    const controller = new AbortController();
                    const timeoutId = setTimeout(() => controller.abort(), 10 * 60 * 1000);

                    fetch('/api/download', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ matchId: matchId }),
                        signal: controller.signal
                    })
                    .then(async res => {
                        clearTimeout(timeoutId);
                        if (!res.ok) {
                            const errorText = await res.text();
                            throw new Error(errorText || res.statusText);
                        }
                        return res.json();
                    })
                    .then(data => {
                        completed++;
                        downloadAllBtn.textContent = `Downloading... (${completed}/${matchesToDownload.length})`;
                        console.log(`Downloaded ${matchId}:`, data);
                        resolve({ matchId, success: true });
                    })
                    .catch(err => {
                        clearTimeout(timeoutId);
                        failed++;
                        const errorMsg = err.name === 'AbortError' 
                            ? 'Request timed out' 
                            : err.message;
                        errors.push(`${matchId}: ${errorMsg}`);
                        downloadAllBtn.textContent = `Downloading... (${completed}/${matchesToDownload.length}, ${failed} failed)`;
                        console.error(`Failed to download ${matchId}:`, err);
                        resolve({ matchId, success: false, error: errorMsg });
                    });
                }, index * 2000);
            });
        });

        Promise.all(downloadPromises)
            .then(results => {
                downloadAllBtn.textContent = originalText;
                downloadAllBtn.disabled = false;

                if (failed > 0) {
                    alert(`Downloaded ${completed} file(s), but ${failed} failed:\n\n${errors.join('\n')}`);
                } else {
                    alert(`Successfully downloaded ${completed} replay file(s)!`);
                }

                if (typeof window.loadReplays === 'function') {
                    setTimeout(() => window.loadReplays(), 1000);
                }

                fetchHistoryBtn.click();
            })
            .catch(err => {
                downloadAllBtn.textContent = originalText;
                downloadAllBtn.disabled = false;
                alert('Error during batch download: ' + err.message);
            });
    }

    function getExistingMatchIds() {
        const replayList = document.getElementById('replay-list');
        if (!replayList) {
            return new Set();
        }
        const checkboxes = replayList.querySelectorAll('input[type="checkbox"]');
        const existingIds = new Set();
        checkboxes.forEach(cb => {
            const matchId = String(cb.value).trim();
            if (matchId) {
                existingIds.add(matchId);
            }
        });
        return existingIds;
    }

    downloadAllBtn.addEventListener('click', downloadAllMatches);
});
