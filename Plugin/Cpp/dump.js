function toggleChildren(btn) {
    const parent = btn.closest('.node');
    parent.classList.toggle('collapsed');
}

function showDetail(details) {
    const overlay = document.getElementById('detailOverlay');
    const content = document.getElementById('detailContent');
    let html = '';
    details.forEach(detail => {
        html += `<div class='detail-block'>
      <p><strong>Filename:</strong> ${detail.filename}</p>
      <p><strong>Source File:</strong> ${detail.sourceFile}</p>
      <p><strong>Source Line:</strong> ${detail.sourceLine}</p>
    </div>`;
    });
    content.innerHTML = html;
    overlay.classList.add('active');
}

function hideDetail() {
    const overlay = document.getElementById('detailOverlay');
    overlay.classList.remove('active');
}

document.addEventListener('contextmenu', (e) => {
    const nodeHeader = e.target.closest('.node-header');
    if (nodeHeader) {
        e.preventDefault();
        const details = JSON.parse(nodeHeader.getAttribute('data-details'));
        showDetail(details);
    }
});

document.addEventListener('dblclick', (e) => {
    const nodeHeader = e.target.closest('.node-header');
    if (nodeHeader) {
        const arrow = nodeHeader.querySelector('.arrow');
        if (arrow) toggleChildren(arrow);
    }
});