import { CenterCanvas } from './CenterCanvas';
import { LeftRail } from './LeftRail';
import { RightRail } from './RightRail';

export function Shell() {
  return (
    <div className="flex h-screen w-screen overflow-hidden bg-slate-50">
      <LeftRail />
      <CenterCanvas />
      <RightRail />
    </div>
  );
}
