import { Route, Routes } from "react-router-dom";
import { Shell } from "@/components/layout/Shell";

export function AppRoutes() {
	return (
		<Routes>
			<Route path="/" element={<Shell />} />
		</Routes>
	);
}
