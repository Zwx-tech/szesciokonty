import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";
import { NameForm } from "./NameForm";

export class NameView extends View {
  private form: NameForm;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("view-center");
    this.form = new NameForm();
    this.form.onContinue = (name) => void this.continue(name);
    this.root.append(this.form);
  }

  onEnter(): void {
    if (this.session.name) this.form.nameValue = this.session.name;
    this.form.error = "";
    requestAnimationFrame(() => this.form.focusField());
  }

  private async continue(name: string): Promise<void> {
    if (this.form.busy) return;
    if (!name) {
      this.form.error = "Enter a name";
      return;
    }
    this.form.busy = true;
    this.form.error = "";
    this.session.setName(name);
    try {
      await this.session.ensureConnected();
      const reconnecting = await this.session.tryReconnect();
      if (!reconnecting) this.router.show("hub");
    } catch {
      this.form.error = "Could not connect";
    } finally {
      this.form.busy = false;
    }
  }
}
