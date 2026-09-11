import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";
import { HubForm } from "./HubForm";

export class HubView extends View {
  private form: HubForm;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("view-center");
    this.form = new HubForm();
    this.form.onCreate = () => void this.create();
    this.form.onJoin = (code) => void this.join(code);
    this.form.onGallery = () => this.router.show("tiles");
    this.root.append(this.form);
  }

  onEnter(): void {
    this.syncError();
  }

  onSessionUpdate(): void {
    this.syncError();
  }

  private syncError(): void {
    this.form.error = this.session.error ?? "";
  }

  private async create(): Promise<void> {
    if (this.form.busy) return;
    this.form.busy = true;
    try {
      await this.session.ensureConnected();
      this.session.createRoom();
    } catch {
      this.session.setError("Could not connect");
    } finally {
      this.form.busy = false;
    }
  }

  private async join(code: string): Promise<void> {
    if (this.form.busy) return;
    if (!code) {
      this.form.error = "Enter a room code";
      return;
    }
    this.form.busy = true;
    try {
      await this.session.ensureConnected();
      this.session.joinRoom(code);
    } catch {
      this.session.setError("Could not connect");
    } finally {
      this.form.busy = false;
    }
  }
}
