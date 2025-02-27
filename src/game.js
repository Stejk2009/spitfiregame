const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');

const plane = new Plane();

function gameLoop() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    plane.draw(ctx);
    // ...existing code...
    requestAnimationFrame(gameLoop);
}

gameLoop();
