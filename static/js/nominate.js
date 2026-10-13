const members = [...document.querySelectorAll('#member-source li')]
    .map(li => ({ username: li.dataset.username, label: li.dataset.label }));
const selected = []; // usernames only
let matches = [];
let active = -1;

const form = document.getElementById('nominate-form');
const field = document.getElementById('nominee-field');
const input = document.getElementById('nominee-input');
const dropdown = document.getElementById('nominee-dropdown');
const errorEl = document.getElementById('nominee-error');

field.addEventListener('click', () => input.focus());

function renderDropdown() {
    const query = input.value.trim().toLowerCase();
    dropdown.replaceChildren();
    if (!query) {
        dropdown.hidden = true;
        return;
    }

    // the label contains first name, last name, and username, so all three are searchable
    matches = members
        .filter(m => !selected.includes(m.username) && m.label.toLowerCase().includes(query))
        .slice(0, 8);
    active = matches.length ? 0 : -1;

    if (matches.length === 0) {
        const none = document.createElement('div');
        none.className = 'list-group-item text-body-secondary';
        none.textContent = 'No matching members';
        dropdown.append(none);
    } else {
        matches.forEach((member, i) => {
            const item = document.createElement('button');
            item.type = 'button';
            item.className = 'list-group-item list-group-item-action' + (i === active ? ' active' : '');
            item.textContent = member.label;
            // mousedown, not click, so the input doesn't lose focus first
            item.addEventListener('mousedown', (e) => {
                e.preventDefault();
                addChip(member);
            });
            dropdown.append(item);
        });
    }
    dropdown.hidden = false;
}

function highlight(index) {
    const items = dropdown.querySelectorAll('button');
    items.forEach((el, i) => el.classList.toggle('active', i === index));
    items[index]?.scrollIntoView({ block: 'nearest' });
}

function addChip(member) {
    if (selected.includes(member.username)) return;
    selected.push(member.username);
    errorEl.textContent = '';

    const chip = document.createElement('span');
    chip.className = 'badge text-bg-primary d-inline-flex align-items-center gap-2 fs-6 fw-normal';
    chip.dataset.username = member.username;
    chip.append(member.label);

    // only the bare username is submitted
    const hidden = document.createElement('input');
    hidden.type = 'hidden';
    hidden.name = 'usernames';
    hidden.value = member.username;

    const remove = document.createElement('button');
    remove.type = 'button';
    remove.className = 'btn-close btn-close-white';
    remove.setAttribute('aria-label', 'Remove ' + member.label);
    remove.addEventListener('click', (e) => {
        e.stopPropagation();
        removeChip(member.username);
    });

    chip.append(hidden, remove);
    field.insertBefore(chip, input);

    input.value = '';
    dropdown.hidden = true;
    input.focus();
}

function removeChip(username) {
    const i = selected.indexOf(username);
    if (i !== -1) selected.splice(i, 1);
    field.querySelector('[data-username="' + CSS.escape(username) + '"]')?.remove();
    input.focus();
}

input.addEventListener('input', renderDropdown);
input.addEventListener('focus', renderDropdown);

input.addEventListener('keydown', (e) => {
    if (e.key === 'ArrowDown' && matches.length) {
        e.preventDefault();
        active = (active + 1) % matches.length;
        highlight(active);
    } else if (e.key === 'ArrowUp' && matches.length) {
        e.preventDefault();
        active = (active - 1 + matches.length) % matches.length;
        highlight(active);
    } else if (e.key === 'Enter') {
        e.preventDefault(); // never submit the form from the text box
        if (!dropdown.hidden && active >= 0) addChip(matches[active]);
    } else if (e.key === 'Escape') {
        dropdown.hidden = true;
    } else if (e.key === 'Backspace' && input.value === '' && selected.length) {
        removeChip(selected[selected.length - 1]);
    }
});

document.addEventListener('click', (e) => {
    if (!field.contains(e.target) && !dropdown.contains(e.target)) {
        dropdown.hidden = true;
    }
});

form.addEventListener('submit', (e) => {
    if (selected.length === 0) {
        e.preventDefault();
        errorEl.textContent = 'Add at least one person to nominate.';
    }
});