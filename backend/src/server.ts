import { app, port } from './presentation/app';

app.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});