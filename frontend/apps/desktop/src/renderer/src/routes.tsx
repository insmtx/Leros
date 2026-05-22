import { Shell } from "@leros/app-ui/components/layout/Shell";
import { Route, Routes } from "react-router-dom";
import logoUrl from "@/assets/logo.svg";

export function AppRoutes() {
	return (
		<Routes>
			<Route path="/" element={<Shell logoSrc={logoUrl} />} />
		</Routes>
	);
}
