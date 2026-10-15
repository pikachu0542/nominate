const grid = document.getElementById('period-grid');

// Data in the order it was received from the server
const cards = [...grid.children];

const statusSelect = document.getElementById('filter-status');
const positionSelect = document.getElementById('filter-position');
const sortSelect = document.getElementById('sort-by');
const countEl = document.getElementById('period-count');
const emptyEl = document.getElementById('period-empty');

const comparators = {
    opens_desc: (a, b) => b.dataset.opens - a.dataset.opens,
    opens_asc: (a, b) => a.dataset.opens - b.dataset.opens,
    closes_desc: (a, b) => b.dataset.closes - a.dataset.closes,
    closes_asc: (a, b) => a.dataset.closes - b.dataset.closes,
    position_asc: (a, b) => a.dataset.position.toLowerCase().localeCompare(b.dataset.position.toLowerCase()),
    status_asc: (a, b) => b.dataset.status.localeCompare(a.dataset.status),
};

function apply() {
    const status = statusSelect.value;
    const position = positionSelect.value;

    const visible = cards.filter(card =>
        (!status || card.dataset.status === status) &&
        (!position || card.dataset.positionId === position)
    );

    cards.forEach(card => { card.hidden = !visible.includes(card); });

    // Array.sort is stable, so ties keep the server's order
    visible.sort(comparators[sortSelect.value]);
    visible.forEach(card => grid.append(card));

    countEl.textContent = 'Showing ' + visible.length + ' of ' + cards.length + ' periods';
    emptyEl.hidden = visible.length > 0;
}

[statusSelect, positionSelect, sortSelect].forEach(el => el.addEventListener('change', apply));

document.getElementById('reset-filters').addEventListener('click', () => {
    statusSelect.value = '';
    positionSelect.value = '';
    sortSelect.value = 'opens_desc';
    apply();
});

apply();