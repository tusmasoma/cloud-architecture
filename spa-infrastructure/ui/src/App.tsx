import { Container, Stack } from "@chakra-ui/react";
import Navbar from "./components/Navbar";
import TodoForm from "./components/TodoForm";
import TodoList from "./components/TodoList";

// export const BASE_URL = import.meta.env.MODE === "development" ? "http://localhost:5000/api" : "/api";
export const BASE_URL = "http://localhost:8080/api";

function App() {
	return (
		<Stack h='100vh'>
			<Navbar />
			<Container>
				<TodoForm />
				<TodoList />
			</Container>
		</Stack>
	);
}

export default App;