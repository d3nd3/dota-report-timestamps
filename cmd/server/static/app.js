document.addEventListener('DOMContentLoaded', () => {
    const replayDirInput = document.getElementById('replay-dir');
    const stratzApiTokenInput = document.getElementById('stratz-api-token');
    const saveConfigBtn = document.getElementById('save-config');
    const replayList = document.getElementById('replay-list');
    const refreshReplaysBtn = document.getElementById('refresh-replays');
    const sortDateBtn = document.getElementById('sort-date');
    const selectAllBtn = document.getElementById('select-all');
    const deselectAllBtn = document.getElementById('deselect-all');
    const deleteSelectedBtn = document.getElementById('delete-selected');
    const startParseBtn = document.getElementById('start-parse');
    const steamIdInput = document.getElementById('steam-id');
    const slotIdInput = document.getElementById('slot-id');
    const progressSection = document.getElementById('progress-section');
    const progressBar = document.getElementById('progress-bar');
    const progressText = document.getElementById('progress-text');
    const resultsSection = document.getElementById('results-section');
    const conclusionDiv = document.getElementById('conclusion');
    const resultsLog = document.getElementById('results-log');
    const historySteamIdInput = document.getElementById('history-steam-id');
    const historyLimitInput = document.getElementById('history-limit');
    const fetchHistoryBtn = document.getElementById('fetch-history');
    const historyResults = document.getElementById('history-results');

    // Load from localStorage
    function loadFromStorage() {
        const savedToken = localStorage.getItem('stratzApiToken');
        const savedSteamId = localStorage.getItem('steamId');
        const savedSlotId = localStorage.getItem('slotId');
        
        if (savedToken) {
            stratzApiTokenInput.value = savedToken;
            console.log('Loaded token from localStorage (length:', savedToken.length, ')');
        }
        if (savedSteamId) steamIdInput.value = savedSteamId;
        if (savedSlotId) slotIdInput.value = savedSlotId;
    }

    // Save to localStorage
    function saveToStorage() {
        if (stratzApiTokenInput.value) localStorage.setItem('stratzApiToken', stratzApiTokenInput.value);
        if (steamIdInput.value) localStorage.setItem('steamId', steamIdInput.value);
        if (slotIdInput.value) localStorage.setItem('slotId', slotIdInput.value);
    }

    // Load from localStorage first (before server config)
    loadFromStorage();

    // Auto-save token to server if it exists in localStorage on page load
    const initialToken = localStorage.getItem('stratzApiToken');
    if (initialToken && initialToken.trim() !== '') {
        setTimeout(() => {
            fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ stratzApiToken: initialToken })
            }).catch(err => console.error('Auto-save on load failed:', err));
        }, 500);
    }

    // Final safety check after a short delay to catch any browser clearing behavior
    setTimeout(() => {
        if (!stratzApiTokenInput.value) {
            const savedToken = localStorage.getItem('stratzApiToken');
            if (savedToken) {
                console.log('Late check: Token field was cleared, restoring from localStorage');
                stratzApiTokenInput.value = savedToken;
            }
        }
    }, 1000);

    // Load Config from server (will override localStorage if present)
    fetch('/api/config')
        .then(res => res.json())
        .then(config => {
            replayDirInput.value = config.replayDir || '';
            // Only update token if server has a non-empty value
            if (config.stratzApiToken && config.stratzApiToken.trim() !== '') {
                console.log('Server has token, updating from server (length:', config.stratzApiToken.length, ')');
                stratzApiTokenInput.value = config.stratzApiToken;
                localStorage.setItem('stratzApiToken', config.stratzApiToken);
            } else {
                console.log('Server token is empty/missing, keeping localStorage value');
            }
            // If server token is empty/missing, keep the localStorage value (already loaded above)
            // Double-check: if token field is empty but localStorage has it, restore it
            if (!stratzApiTokenInput.value) {
                const savedToken = localStorage.getItem('stratzApiToken');
                if (savedToken) {
                    console.log('Token field was cleared, restoring from localStorage');
                    stratzApiTokenInput.value = savedToken;
                }
            }
            loadReplays();
        })
        .catch(() => {
            // If server config fails, values from loadFromStorage() above will be used
            // Double-check token is still there
            if (!stratzApiTokenInput.value) {
                const savedToken = localStorage.getItem('stratzApiToken');
                if (savedToken) {
                    stratzApiTokenInput.value = savedToken;
                }
            }
        });

    // Auto-save token when it changes (with debounce)
    let saveTokenTimeout;
    stratzApiTokenInput.addEventListener('input', () => {
        const token = stratzApiTokenInput.value.trim();
        localStorage.setItem('stratzApiToken', token);
        
        // Auto-save to server after user stops typing (1 second delay)
        clearTimeout(saveTokenTimeout);
        saveTokenTimeout = setTimeout(() => {
            if (token) {
                fetch('/api/config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ stratzApiToken: token })
                }).catch(err => console.error('Auto-save failed:', err));
            }
        }, 1000);
    });
    stratzApiTokenInput.addEventListener('change', () => {
        const token = stratzApiTokenInput.value.trim();
        localStorage.setItem('stratzApiToken', token);
        if (token) {
            fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ stratzApiToken: token })
            }).catch(err => console.error('Auto-save failed:', err));
        }
    });
    steamIdInput.addEventListener('input', () => {
        if (steamIdInput.value) localStorage.setItem('steamId', steamIdInput.value);
    });
    slotIdInput.addEventListener('input', () => {
        if (slotIdInput.value) localStorage.setItem('slotId', slotIdInput.value);
    });

    // Save Config
    saveConfigBtn.addEventListener('click', () => {
        const newDir = replayDirInput.value;
        const newToken = stratzApiTokenInput.value.trim();
        console.log('Saving config - Token length:', newToken.length);
        
        // Save to localStorage immediately
        if (newToken) {
            localStorage.setItem('stratzApiToken', newToken);
            console.log('Saved token to localStorage');
        } else {
            console.warn('Token is empty when saving!');
        }
        saveToStorage();
        
        const payload = { replayDir: newDir };
        if (newToken) {
            payload.stratzApiToken = newToken;
        }
        
        fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        })
            .then(res => res.json())
            .then(config => {
                console.log('Server response - Token length:', config.stratzApiToken ? config.stratzApiToken.length : 0);
                // Ensure localStorage is updated with what server saved
                if (config.stratzApiToken) {
                    localStorage.setItem('stratzApiToken', config.stratzApiToken);
                    stratzApiTokenInput.value = config.stratzApiToken;
                }
                alert('Configuration saved!');
                loadReplays();
            })
            .catch(err => {
                console.error('Error saving config:', err);
                alert('Error saving config: ' + err);
            });
    });

    let currentReplays = [];
    let sortOrder = 'desc'; // 'desc' (newest first) or 'asc' (oldest first)

    // Load Replays
    function loadReplays() {
        replayList.innerHTML = '<p class="loading">Loading replays...</p>';
        fetch('/api/replays')
            .then(res => res.json())
            .then(replays => {
                currentReplays = replays || [];
                renderReplayList();
            })
            .catch(err => {
                replayList.innerHTML = `<p style="color:red">Error loading replays: ${err}</p>`;
            });
    }
    
    function renderReplayList() {
        replayList.innerHTML = '';
        if (!currentReplays || currentReplays.length === 0) {
            replayList.innerHTML = '<p>No .dem files found in directory.</p>';
            return;
        }

        // Sort replays
        currentReplays.sort((a, b) => {
            const dateA = a.date ? new Date(a.date) : new Date(0);
            const dateB = b.date ? new Date(b.date) : new Date(0);
            return sortOrder === 'desc' ? dateB - dateA : dateA - dateB;
        });

        currentReplays.forEach(replay => {
            const div = document.createElement('div');
            div.className = 'replay-item';
            const fileName = replay.fileName || replay;
            const id = fileName.replace('.dem', '');
            const date = replay.date ? new Date(replay.date).toLocaleString() : '';
            const dateDisplay = date ? ` <span class="date-display">(${date})</span>` : '';
            div.innerHTML = `
                <input type="checkbox" value="${id}" id="replay-${id}">
                <label for="replay-${id}">${fileName}${dateDisplay}</label>
            `;
            replayList.appendChild(div);
        });
    }

    sortDateBtn.addEventListener('click', () => {
        sortOrder = sortOrder === 'desc' ? 'asc' : 'desc';
        sortDateBtn.textContent = sortOrder === 'desc' ? 'Sort: Newest' : 'Sort: Oldest';
        renderReplayList();
    });
    
    window.loadReplays = loadReplays;

    function deleteSelectedReplays() {
        const checkboxes = document.querySelectorAll('.replay-item input[type="checkbox"]:checked');
        const selectedIds = Array.from(checkboxes).map(cb => cb.value);
        
        if (selectedIds.length === 0) {
            alert('Please select at least one replay to delete.');
            return;
        }
        
        const fileNames = selectedIds.map(id => id + '.dem').join(', ');
        if (!confirm(`Are you sure you want to delete ${selectedIds.length} replay file(s)?\n\n${fileNames}\n\nThis cannot be undone.`)) {
            return;
        }
        
        const originalText = deleteSelectedBtn.textContent;
        deleteSelectedBtn.disabled = true;
        deleteSelectedBtn.textContent = 'Deleting...';
        
        const deletePromises = selectedIds.map(matchId => 
            fetch('/api/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ matchId: matchId })
            }).then(async res => {
                if (!res.ok) {
                    const errorText = await res.text();
                    throw new Error(`Failed to delete ${matchId}: ${errorText || res.statusText}`);
                }
                return res.json();
            })
        );
        
        Promise.allSettled(deletePromises)
            .then(results => {
                const successful = results.filter(r => r.status === 'fulfilled').length;
                const failed = results.filter(r => r.status === 'rejected').length;
                
                if (failed > 0) {
                    const errors = results
                        .filter(r => r.status === 'rejected')
                        .map(r => r.reason.message)
                        .join('\n');
                    alert(`Deleted ${successful} file(s), but ${failed} failed:\n\n${errors}`);
                } else {
                    console.log(`Successfully deleted ${successful} replay file(s)`);
                }
                
                deleteSelectedBtn.textContent = originalText;
                deleteSelectedBtn.disabled = false;
                loadReplays();
            })
            .catch(err => {
                deleteSelectedBtn.textContent = originalText;
                deleteSelectedBtn.disabled = false;
                alert('Error deleting replays: ' + err.message);
            });
    }

    refreshReplaysBtn.addEventListener('click', loadReplays);

    selectAllBtn.addEventListener('click', () => {
        document.querySelectorAll('.replay-item input[type="checkbox"]').forEach(cb => cb.checked = true);
    });

    deselectAllBtn.addEventListener('click', () => {
        document.querySelectorAll('.replay-item input[type="checkbox"]').forEach(cb => cb.checked = false);
    });

    deleteSelectedBtn.addEventListener('click', deleteSelectedReplays);

    // Parse Logic
    startParseBtn.addEventListener('click', async () => {
        const selectedCheckboxes = document.querySelectorAll('.replay-item input[type="checkbox"]:checked');
        const selectedIds = Array.from(selectedCheckboxes).map(cb => cb.value);

        if (selectedIds.length === 0) {
            alert('Please select at least one replay.');
            return;
        }

        // Use string for Steam ID to avoid precision loss, convert to number for JSON
        const steamIdStr = steamIdInput.value ? steamIdInput.value.trim() : '';
        // Keep as string
        const steamIdToSend = steamIdStr || "0";
        const slotId = slotIdInput.value ? parseInt(slotIdInput.value) : -1;
        saveToStorage();

        // Reset UI
        progressSection.classList.remove('hidden');
        resultsSection.classList.add('hidden');
        resultsLog.innerHTML = '';
        conclusionDiv.innerHTML = '';
        progressBar.style.width = '0%';

        let totalTeamReports = 0;
        let totalEnemyReports = 0;
        let totalConfirmedTeamReports = 0;
        let totalConfirmedEnemyReports = 0;
        let processedCount = 0;

        for (const matchId of selectedIds) {
            progressText.textContent = `Processing ${processedCount + 1} / ${selectedIds.length}: ${matchId}`;

            try {
                const res = await fetch('/api/parse', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        matchId: matchId,
                        reportedSlot: slotId,
                        reportedSteamId: steamIdToSend
                    })
                });

                if (!res.ok) throw new Error(await res.text());

                const result = await res.json();

                // Log results
                const logEntry = document.createElement('div');
                logEntry.className = 'log-entry';
                
                let headerHtml = `<div class="log-header"><strong>Match ${result.MatchID}</strong></div>`;
                headerHtml += `<div class="log-summary">Reports: ${result.TeamReports} (Friendly) / ${result.EnemyReports} (Enemy)</div>`;
                headerHtml += `<div class="log-summary">Confirmed: ${result.ConfirmedTeamReports || 0} (Friendly) / ${result.ConfirmedEnemyReports || 0} (Enemy)</div>`;
                
                let reportsHtml = '';
                if (result.Reports) {
                    reportsHtml = '<div class="log-reports">';
                    result.Reports.forEach(report => {
                        let reportText = `_REPORT_ (${report.Time}) from _${report.Team}_: ${report.Name} (Slot ${report.Slot})`;
                        let className = "report-item";
                        if (report.Confirmed) {
                            reportText += ` [CONFIRMED +${report.ConfirmationDelayMs}ms]`;
                            className += " confirmed";
                        } else {
                            reportText += ` [UNCONFIRMED]`;
                        }
                        reportsHtml += `<div class="${className}">${reportText}</div>`;
                    });
                    reportsHtml += '</div>';
                }
                logEntry.innerHTML = headerHtml + reportsHtml;
                resultsLog.appendChild(logEntry);

                totalTeamReports += result.TeamReports;
                totalEnemyReports += result.EnemyReports;
                totalConfirmedTeamReports += (result.ConfirmedTeamReports || 0);
                totalConfirmedEnemyReports += (result.ConfirmedEnemyReports || 0);

            } catch (err) {
                const errorEntry = document.createElement('div');
                errorEntry.style.color = 'red';
                errorEntry.className = 'log-entry';
                let errorMsg = err.message;
                if (errorMsg.includes('no such file or directory') || errorMsg.includes('Could not open replay file')) {
                    errorMsg = `Replay file not found: ${matchId}.dem - Make sure the replay file exists in your replay directory. You may need to download it from Dota 2 first.`;
                }
                errorEntry.innerHTML = `<strong style="color:red">Error processing ${matchId}:</strong> ${errorMsg}`;
                resultsLog.appendChild(errorEntry);
            }

            processedCount++;
            progressBar.style.width = `${(processedCount / selectedIds.length) * 100}%`;
        }

        // Conclusion
        resultsSection.classList.remove('hidden');
        conclusionDiv.innerHTML = `
            <h3>Analysis Complete</h3>
            <p>Processed ${processedCount} replays.</p>
            <p><strong>Total Team Reports:</strong> ${totalTeamReports} (Confirmed: ${totalConfirmedTeamReports})</p>
            <p><strong>Total Enemy Reports:</strong> ${totalEnemyReports} (Confirmed: ${totalConfirmedEnemyReports})</p>
            <p><em>${totalTeamReports + totalEnemyReports > 0 ? "You have been naughty!" : "Clean record! Well played."}</em></p>
        `;
    });
});
