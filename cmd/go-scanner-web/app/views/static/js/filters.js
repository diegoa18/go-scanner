function filterResults() {
    const showClosed = document.getElementById('showClosed').checked;
    const minConfidence = document.getElementById('minConfidence').value;
    const rows = document.getElementsByClassName('result-row');

    for (let row of rows) {
        const status = row.getAttribute('data-status');
        const confidence = row.getAttribute('data-confidence');
        let visible = true;

        // filtro en base del estado
        if (!showClosed && status === 'CLOSED') {
            visible = false;
        }

        //filtro en base a confidence
        if (visible && minConfidence !== 'all') {
            //high > medium > low/unknown
            if (minConfidence === 'high' && confidence !== 'high') visible = false;
            if (minConfidence === 'medium' && (confidence === 'low' || confidence === 'unknown')) visible = false;
        }

        row.style.display = visible ? '' : 'none';
    }
}

//listeners cuando DOM esta listo
document.addEventListener('DOMContentLoaded', function () {
    //se inicializa al momento de existir elementos
    const showClosed = document.getElementById('showClosed');

    if (showClosed) {
        filterResults();
    }
});
