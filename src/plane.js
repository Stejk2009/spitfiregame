class Plane {
    constructor() {
        this.image = new Image();
        this.image.src = 'assets/images/plane.png'; // Ensure this path is correct
        this.x = 0;
        this.y = 0;
    }

    draw(ctx) {
        ctx.drawImage(this.image, this.x, this.y);
    }
}
